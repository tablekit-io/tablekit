package oauth

import (
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"core/services"
	"core/services/config"
	"core/services/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIssuer(t *testing.T, configService *config.Config) *Issuer {
	t.Helper()
	storageService, err := store.New(t.TempDir())
	require.NoError(t, err)
	issuer, err := NewIssuer(&services.Services{Config: configService, Store: storageService})
	require.NoError(t, err)
	return issuer
}

func testConfig() *config.Config {
	return &config.Config{
		PublicBaseURL: "http://localhost:8080",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	issuer := newIssuer(t, testConfig())

	token, err := issuer.IssueAccess("client-1", "chain-1", "mcp")
	require.NoError(t, err)

	claims, err := issuer.VerifyAccess(token)
	require.NoError(t, err)
	assert.Equal(t, "client-1", claims.CID)
	assert.Equal(t, "chain-1", claims.Chain)
	assert.Equal(t, "mcp", claims.Scope)
	assert.Equal(t, "user:owner", claims.Subject)
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	issuer := newIssuer(t, testConfig())

	token, iat, err := issuer.IssueRefresh("client-1", "chain-1", "mcp")
	require.NoError(t, err)
	assert.False(t, iat.IsZero())

	claims, err := issuer.VerifyRefresh(token)
	require.NoError(t, err)
	assert.Equal(t, "chain-1", claims.Chain)
}

func TestAudienceSeparation(t *testing.T) {
	issuer := newIssuer(t, testConfig())

	access, err := issuer.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)
	refresh, _, err := issuer.IssueRefresh("c", "ch", "mcp")
	require.NoError(t, err)

	// An access token must not validate as a refresh token, and vice versa.
	_, err = issuer.VerifyRefresh(access)
	assert.Error(t, err)
	_, err = issuer.VerifyAccess(refresh)
	assert.Error(t, err)
}

func TestExpiredTokenRejected(t *testing.T) {
	configService := testConfig()
	configService.AccessTTL = -time.Minute // already expired at issue time
	issuer := newIssuer(t, configService)

	token, err := issuer.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)

	_, err = issuer.VerifyAccess(token)
	assert.Error(t, err)
}

func TestWrongKeyRejected(t *testing.T) {
	configService := testConfig()
	issuer := newIssuer(t, configService)
	other := newIssuer(t, configService) // different temp dir → different signing key

	token, err := issuer.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)

	_, err = other.VerifyAccess(token)
	assert.Error(t, err)
}

func TestTamperedTokenRejected(t *testing.T) {
	issuer := newIssuer(t, testConfig())
	token, err := issuer.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)

	// Corrupt a character in the payload (just past the first '.'), which
	// invalidates the signature. (Flipping the last signature char can be a
	// no-op: the final base64url char of a 32-byte HMAC has unused low bits.)
	dot := strings.IndexByte(token, '.')
	require.Positive(t, dot)
	i := dot + 1
	tampered := token[:i] + flip(token[i]) + token[i+1:]
	require.NotEqual(t, token, tampered)

	_, err = issuer.VerifyAccess(tampered)
	assert.Error(t, err)
}

func flip(b byte) string {
	if b == 'A' {
		return "B"
	}
	return "A"
}

// newStore is a small helper for the env-key tests that need the store dir.
func newStore(t *testing.T) (*store.Store, string) {
	t.Helper()
	directory := t.TempDir()
	storageService, err := store.New(directory)
	require.NoError(t, err)
	return storageService, directory
}

func TestEnvSigningKeyIsSharedAcrossInstances(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	configService := testConfig()
	configService.SigningKey = base64.StdEncoding.EncodeToString(key)

	// Two issuers, different stores, same external key → cross-verify.
	storageServiceA, _ := newStore(t)
	storageServiceB, _ := newStore(t)
	issuerA, err := NewIssuer(&services.Services{Config: configService, Store: storageServiceA})
	require.NoError(t, err)
	issuerB, err := NewIssuer(&services.Services{Config: configService, Store: storageServiceB})
	require.NoError(t, err)

	token, err := issuerA.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)
	claims, err := issuerB.VerifyAccess(token)
	require.NoError(t, err)
	assert.Equal(t, "ch", claims.Chain)
}

func TestEnvSigningKeyShortIsPadded(t *testing.T) {
	configService := testConfig()
	configService.SigningKey = base64.StdEncoding.EncodeToString([]byte("short-key"))

	storageService, _ := newStore(t)
	issuer, err := NewIssuer(&services.Services{Config: configService, Store: storageService})
	require.NoError(t, err)

	token, err := issuer.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)
	_, err = issuer.VerifyAccess(token)
	require.NoError(t, err)
}

func TestEnvSigningKeyInvalidRejected(t *testing.T) {
	configService := testConfig()
	configService.SigningKey = "!!! not base64 !!!"

	storageService, _ := newStore(t)
	_, err := NewIssuer(&services.Services{Config: configService, Store: storageService})
	assert.Error(t, err)
}

func TestEnvSigningKeyDoesNotWriteFile(t *testing.T) {
	configService := testConfig()
	configService.SigningKey = base64.StdEncoding.EncodeToString(make([]byte, 32))

	storageService, directory := newStore(t)
	_, err := NewIssuer(&services.Services{Config: configService, Store: storageService})
	require.NoError(t, err)

	// Env key takes precedence; the store's key file is never created.
	assert.NoFileExists(t, filepath.Join(directory, "signing.key"))
}

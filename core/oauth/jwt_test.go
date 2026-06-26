package oauth

import (
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"core/config"
	"core/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIssuer(t *testing.T, cfg *config.Config) *Issuer {
	t.Helper()
	st, err := store.New(t.TempDir())
	require.NoError(t, err)
	iss, err := NewIssuer(cfg, st)
	require.NoError(t, err)
	return iss
}

func testConfig() *config.Config {
	return &config.Config{
		PublicBaseURL: "http://localhost:8080",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	iss := newIssuer(t, testConfig())

	tok, err := iss.IssueAccess("client-1", "chain-1", "mcp")
	require.NoError(t, err)

	claims, err := iss.VerifyAccess(tok)
	require.NoError(t, err)
	assert.Equal(t, "client-1", claims.CID)
	assert.Equal(t, "chain-1", claims.Chain)
	assert.Equal(t, "mcp", claims.Scope)
	assert.Equal(t, "user:owner", claims.Subject)
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	iss := newIssuer(t, testConfig())

	tok, iat, err := iss.IssueRefresh("client-1", "chain-1", "mcp")
	require.NoError(t, err)
	assert.False(t, iat.IsZero())

	claims, err := iss.VerifyRefresh(tok)
	require.NoError(t, err)
	assert.Equal(t, "chain-1", claims.Chain)
}

func TestAudienceSeparation(t *testing.T) {
	iss := newIssuer(t, testConfig())

	access, err := iss.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)
	refresh, _, err := iss.IssueRefresh("c", "ch", "mcp")
	require.NoError(t, err)

	// An access token must not validate as a refresh token, and vice versa.
	_, err = iss.VerifyRefresh(access)
	assert.Error(t, err)
	_, err = iss.VerifyAccess(refresh)
	assert.Error(t, err)
}

func TestExpiredTokenRejected(t *testing.T) {
	cfg := testConfig()
	cfg.AccessTTL = -time.Minute // already expired at issue time
	iss := newIssuer(t, cfg)

	tok, err := iss.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)

	_, err = iss.VerifyAccess(tok)
	assert.Error(t, err)
}

func TestWrongKeyRejected(t *testing.T) {
	cfg := testConfig()
	iss := newIssuer(t, cfg)
	other := newIssuer(t, cfg) // different temp dir → different signing key

	tok, err := iss.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)

	_, err = other.VerifyAccess(tok)
	assert.Error(t, err)
}

func TestTamperedTokenRejected(t *testing.T) {
	iss := newIssuer(t, testConfig())
	tok, err := iss.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)

	// Corrupt a character in the payload (just past the first '.'), which
	// invalidates the signature. (Flipping the last signature char can be a
	// no-op: the final base64url char of a 32-byte HMAC has unused low bits.)
	dot := strings.IndexByte(tok, '.')
	require.Positive(t, dot)
	i := dot + 1
	tampered := tok[:i] + flip(tok[i]) + tok[i+1:]
	require.NotEqual(t, tok, tampered)

	_, err = iss.VerifyAccess(tampered)
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
	dir := t.TempDir()
	st, err := store.New(dir)
	require.NoError(t, err)
	return st, dir
}

func TestEnvSigningKeyIsSharedAcrossInstances(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	cfg := testConfig()
	cfg.SigningKey = base64.StdEncoding.EncodeToString(key)

	// Two issuers, different stores, same external key → cross-verify.
	stA, _ := newStore(t)
	stB, _ := newStore(t)
	issA, err := NewIssuer(cfg, stA)
	require.NoError(t, err)
	issB, err := NewIssuer(cfg, stB)
	require.NoError(t, err)

	tok, err := issA.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)
	claims, err := issB.VerifyAccess(tok)
	require.NoError(t, err)
	assert.Equal(t, "ch", claims.Chain)
}

func TestEnvSigningKeyShortIsPadded(t *testing.T) {
	cfg := testConfig()
	cfg.SigningKey = base64.StdEncoding.EncodeToString([]byte("short-key"))

	st, _ := newStore(t)
	iss, err := NewIssuer(cfg, st)
	require.NoError(t, err)

	tok, err := iss.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)
	_, err = iss.VerifyAccess(tok)
	require.NoError(t, err)
}

func TestEnvSigningKeyInvalidRejected(t *testing.T) {
	cfg := testConfig()
	cfg.SigningKey = "!!! not base64 !!!"

	st, _ := newStore(t)
	_, err := NewIssuer(cfg, st)
	assert.Error(t, err)
}

func TestEnvSigningKeyDoesNotWriteFile(t *testing.T) {
	cfg := testConfig()
	cfg.SigningKey = base64.StdEncoding.EncodeToString(make([]byte, 32))

	st, dir := newStore(t)
	_, err := NewIssuer(cfg, st)
	require.NoError(t, err)

	// Env key takes precedence; the store's key file is never created.
	assert.NoFileExists(t, filepath.Join(dir, "signing.key"))
}

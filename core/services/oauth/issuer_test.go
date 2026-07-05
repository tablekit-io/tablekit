package oauth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"os"
	"strings"
	"testing"
	"time"

	"core/db/dbtest"
	"core/services/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available), so each test gets an isolated migrated database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

func newIssuer(t *testing.T, configService *config.Config) *Issuer {
	t.Helper()
	// Give the issuer a signing key if the caller didn't set one. A fresh config
	// gets a fresh random key, so two issuers built from separate testConfig()
	// values use different keys (see TestWrongKeyRejected).
	if configService.SigningKey == "" {
		configService.SigningKey = randomSigningKey(t)
	}
	issuer, err := NewIssuer(configService)
	require.NoError(t, err)
	return issuer
}

// randomSigningKey returns a fresh base64 32-byte HS256 key.
func randomSigningKey(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(key)
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

func TestBearerTokenRoundTrip(t *testing.T) {
	issuer := newIssuer(t, testConfig())

	token, expiresAt, err := issuer.IssueBearer("client-1", "tok-1")
	require.NoError(t, err)

	// Valid roughly 6 calendar months out.
	assert.WithinDuration(t, time.Now().AddDate(0, 6, 0), expiresAt, time.Minute)

	claims, err := issuer.VerifyBearer(token)
	require.NoError(t, err)
	assert.Equal(t, "client-1", claims.CID)
	assert.Equal(t, "tok-1", claims.ID)
	assert.Equal(t, "mcp", claims.Scope)
	// exp is truncated to the JWT time precision (microsecond), so compare with
	// a small tolerance rather than exact equality.
	assert.WithinDuration(t, expiresAt, claims.ExpiresAt.Time, time.Millisecond)
}

func TestBearerAudienceSeparation(t *testing.T) {
	issuer := newIssuer(t, testConfig())

	bearer, _, err := issuer.IssueBearer("c", "t")
	require.NoError(t, err)
	access, err := issuer.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)

	// A bearer token must not pass as an access token (this is what stops a
	// caller from dropping the prefix to bypass the revocation check), and an
	// access token must not pass as a bearer token.
	_, err = issuer.VerifyAccess(bearer)
	assert.Error(t, err)
	_, err = issuer.VerifyBearer(access)
	assert.Error(t, err)
}

func TestExportTokenRoundTrip(t *testing.T) {
	issuer := newIssuer(t, testConfig())

	token, err := issuer.IssueExport("query-key-1")
	require.NoError(t, err)

	claims, err := issuer.VerifyExport(token)
	require.NoError(t, err)
	assert.Equal(t, "query-key-1", claims.QK)
}

func TestExportAudienceSeparation(t *testing.T) {
	issuer := newIssuer(t, testConfig())

	export, err := issuer.IssueExport("qk")
	require.NoError(t, err)
	access, err := issuer.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)

	// An export token is only honoured by the export endpoint, never on the
	// MCP/OAuth paths, and an access token can't be used as an export link.
	_, err = issuer.VerifyAccess(export)
	assert.Error(t, err)
	_, err = issuer.VerifyExport(access)
	assert.Error(t, err)
}

func TestBearerTokenID(t *testing.T) {
	issuer := newIssuer(t, testConfig())

	token, _, err := issuer.IssueBearer("c", "tok-42")
	require.NoError(t, err)

	// jti is readable without verification (used by token:revoke).
	id, err := BearerTokenID(token)
	require.NoError(t, err)
	assert.Equal(t, "tok-42", id)

	_, err = BearerTokenID("not.a.jwt")
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
	issuer := newIssuer(t, testConfig())
	other := newIssuer(t, testConfig()) // separate config → fresh random key

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

// openTestDB returns a fresh migrated Postgres database, dropped at test end.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return dbtest.New(t)
}

func TestEnvSigningKeyIsSharedAcrossInstances(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	configService := testConfig()
	configService.SigningKey = base64.StdEncoding.EncodeToString(key)

	// Two issuers, same external key → cross-verify.
	issuerA, err := NewIssuer(configService)
	require.NoError(t, err)
	issuerB, err := NewIssuer(configService)
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

	issuer, err := NewIssuer(configService)
	require.NoError(t, err)

	token, err := issuer.IssueAccess("c", "ch", "mcp")
	require.NoError(t, err)
	_, err = issuer.VerifyAccess(token)
	require.NoError(t, err)
}

func TestEnvSigningKeyInvalidRejected(t *testing.T) {
	configService := testConfig()
	configService.SigningKey = "!!! not base64 !!!"

	_, err := NewIssuer(configService)
	assert.Error(t, err)
}

func TestMissingSigningKeyRejected(t *testing.T) {
	configService := testConfig() // SigningKey empty

	_, err := NewIssuer(configService)
	assert.Error(t, err, "a missing SIGNING_KEY must fail issuer construction")
}

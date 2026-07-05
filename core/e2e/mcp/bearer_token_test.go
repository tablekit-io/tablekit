package mcp

import (
	"context"
	"strings"
	"testing"

	"core/e2e/harness"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBearerTokenGrantsMCPAccess: a CLI-minted bearer token reaches /mcp without
// any OAuth flow, alongside the existing OAuth path.
func TestBearerTokenGrantsMCPAccess(t *testing.T) {
	server := harness.StartServer(t)
	_, token := harness.GenerateToken(t, server)

	clientSession, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	list, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	require.NotEmpty(t, list.Tools)
}

// TestBearerTokenCoexistsWithOAuth: both auth modes work against one server.
func TestBearerTokenCoexistsWithOAuth(t *testing.T) {
	server := harness.StartServer(t)

	// OAuth access token (the default "once" mode admits this client).
	_, tokens := harness.FullHandshake(t, server.AppURL)
	oauthSession, err := harness.Connect(t, server.AppURL, harness.BearerClient(tokens["access_token"].(string)))
	require.NoError(t, err)
	t.Cleanup(func() { _ = oauthSession.Close() })

	// CLI bearer token works too, no pairing/handshake needed.
	_, token := harness.GenerateToken(t, server)
	bearerSession, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = bearerSession.Close() })
}

// TestBearerTokenRevoked: once revoked, the token is rejected at /mcp. The raw
// JWT without the prefix must not work either (it would otherwise dodge the
// revocation check).
func TestBearerTokenRevoked(t *testing.T) {
	server := harness.StartServer(t)
	tokenID, token := harness.GenerateToken(t, server)

	// Works before revocation.
	session, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	_ = session.Close()

	// The raw JWT (prefix stripped) is not a valid access token: wrong audience.
	rawJWT := strings.TrimPrefix(token, harness.TokenPrefix)
	_, err = harness.Connect(t, server.AppURL, harness.BearerClient(rawJWT))
	assert.Error(t, err, "raw bearer JWT must not authenticate on the OAuth path")

	// Revoke by id, then the token is rejected (the server checks the token's
	// row in the database per request).
	harness.RunCLI(t, server, "pairing", "token:revoke", tokenID)
	_, err = harness.Connect(t, server.AppURL, harness.BearerClient(token))
	assert.Error(t, err, "revoked bearer token must be rejected")
}

// TestBearerTokenRevokeByToken: revocation accepts the full token, not just the id.
func TestBearerTokenRevokeByToken(t *testing.T) {
	server := harness.StartServer(t)
	_, token := harness.GenerateToken(t, server)

	harness.RunCLI(t, server, "pairing", "token:revoke", token)
	_, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	assert.Error(t, err, "revoked-by-token bearer token must be rejected")
}

package e2e

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tokenPrefix mirrors oauth.TokenPrefix; the e2e suite treats the binary as a
// black box, so it asserts on the wire format rather than importing the const.
const tokenPrefix = "tablekit_pat_"

// generateToken runs `pairing token:generate` against the server's data dir and
// PUBLIC_BASE_URL (the latter must match the server so the minted token's issuer
// claim verifies), returning the printed token id and the full bearer token.
func generateToken(t *testing.T, server server) (tokenID, token string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), binPath, "pairing", "token:generate")
	cmd.Env = append(os.Environ(), "DATA_DIR="+server.dataDir, "PUBLIC_BASE_URL="+server.appURL)
	out, err := cmd.Output()
	require.NoError(t, err, "token:generate failed")

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "token id:"):
			tokenID = strings.TrimSpace(strings.TrimPrefix(line, "token id:"))
		case strings.HasPrefix(line, tokenPrefix):
			token = line
		}
	}
	require.NotEmpty(t, tokenID, "token id not found in output: %s", out)
	require.NotEmpty(t, token, "bearer token not found in output: %s", out)
	return tokenID, token
}

// TestBearerTokenGrantsMCPAccess: a CLI-minted bearer token reaches /mcp without
// any OAuth flow, alongside the existing OAuth path.
func TestBearerTokenGrantsMCPAccess(t *testing.T) {
	server := startServer(t)
	_, token := generateToken(t, server)

	clientSession, err := connect(t, server.appURL, bearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	list, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	require.NotEmpty(t, list.Tools)
}

// TestBearerTokenCoexistsWithOAuth: both auth modes work against one server.
func TestBearerTokenCoexistsWithOAuth(t *testing.T) {
	server := startServer(t)

	// OAuth access token (the default "once" mode admits this client).
	_, tokens := fullHandshake(t, server.appURL)
	oauthSession, err := connect(t, server.appURL, bearerClient(tokens["access_token"].(string)))
	require.NoError(t, err)
	t.Cleanup(func() { _ = oauthSession.Close() })

	// CLI bearer token works too, no pairing/handshake needed.
	_, token := generateToken(t, server)
	bearerSession, err := connect(t, server.appURL, bearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = bearerSession.Close() })
}

// TestBearerTokenRevoked: once revoked, the token is rejected at /mcp. The raw
// JWT without the prefix must not work either (it would otherwise dodge the
// revocation check).
func TestBearerTokenRevoked(t *testing.T) {
	server := startServer(t)
	tokenID, token := generateToken(t, server)

	// Works before revocation.
	session, err := connect(t, server.appURL, bearerClient(token))
	require.NoError(t, err)
	_ = session.Close()

	// The raw JWT (prefix stripped) is not a valid access token: wrong audience.
	rawJWT := strings.TrimPrefix(token, tokenPrefix)
	_, err = connect(t, server.appURL, bearerClient(rawJWT))
	assert.Error(t, err, "raw bearer JWT must not authenticate on the OAuth path")

	// Revoke by id, then the token is rejected (the server reads tokens.json per request).
	runCLI(t, server.dataDir, "pairing", "token:revoke", tokenID)
	_, err = connect(t, server.appURL, bearerClient(token))
	assert.Error(t, err, "revoked bearer token must be rejected")
}

// TestBearerTokenRevokeByToken: revocation accepts the full token, not just the id.
func TestBearerTokenRevokeByToken(t *testing.T) {
	server := startServer(t)
	_, token := generateToken(t, server)

	runCLI(t, server.dataDir, "pairing", "token:revoke", token)
	_, err := connect(t, server.appURL, bearerClient(token))
	assert.Error(t, err, "revoked-by-token bearer token must be rejected")
}

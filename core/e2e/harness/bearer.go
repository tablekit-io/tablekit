package harness

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TokenPrefix mirrors oauth.TokenPrefix; the e2e suite treats the binary as a
// black box, so it asserts on the wire format rather than importing the const.
const TokenPrefix = "tablekit_pat_"

// BearerClient returns an HTTP client that injects a bearer token on every request.
func BearerClient(token string) *http.Client {
	return &http.Client{Transport: bearerRoundTripper{token: token, base: http.DefaultTransport}}
}

type bearerRoundTripper struct {
	token string
	base  http.RoundTripper
}

func (b bearerRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	request = request.Clone(request.Context())
	request.Header.Set("Authorization", "Bearer "+b.token)
	return b.base.RoundTrip(request)
}

// GenerateToken runs `pairing token:generate` against the server's data dir and
// PUBLIC_BASE_URL (the latter must match the server so the minted token's issuer
// claim verifies), returning the printed token id and the full bearer token.
func GenerateToken(t *testing.T, server Server) (tokenID, token string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), ensureBinary(t), "pairing", "token:generate")
	cmd.Env = append(os.Environ(), "DATA_DIR="+server.DataDir, "PUBLIC_BASE_URL="+server.AppURL)
	cmd.Env = append(cmd.Env, server.DBEnv...)
	out, err := cmd.Output()
	require.NoError(t, err, "token:generate failed")

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "token id:"):
			tokenID = strings.TrimSpace(strings.TrimPrefix(line, "token id:"))
		case strings.HasPrefix(line, TokenPrefix):
			token = line
		}
	}
	require.NotEmpty(t, tokenID, "token id not found in output: %s", out)
	require.NotEmpty(t, token, "bearer token not found in output: %s", out)
	return tokenID, token
}

// RunCLI runs the binary with a subcommand against the server's data dir and its
// own state database (so pairing/token changes hit the same DB the server uses).
func RunCLI(t *testing.T, server Server, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), ensureBinary(t), args...)
	cmd.Env = append(os.Environ(), "DATA_DIR="+server.DataDir)
	cmd.Env = append(cmd.Env, server.DBEnv...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "cli %v failed: %s", args, out)
}

// RunCLIOutput runs the binary like RunCLI but returns its stdout, for commands
// whose output the test needs to parse.
func RunCLIOutput(t *testing.T, server Server, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), ensureBinary(t), args...)
	cmd.Env = append(os.Environ(), "DATA_DIR="+server.DataDir)
	cmd.Env = append(cmd.Env, server.DBEnv...)
	out, err := cmd.Output()
	require.NoError(t, err, "cli %v failed", args)
	return string(out)
}

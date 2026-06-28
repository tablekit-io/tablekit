package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"path/filepath"
	"testing"

	"core/e2e/harness"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvSigningKey drives the externally-provided key path through the real
// binary: with SIGNING_KEY set, the full OAuth + MCP flow works and the store
// never writes its own signing.key.
func TestEnvSigningKey(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	server := harness.StartServerEnv(t, "SIGNING_KEY="+base64.StdEncoding.EncodeToString(key))

	_, tokens := harness.FullHandshake(t, server.AppURL)
	token := tokens["access_token"].(string)

	clientSession, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "hello_world",
		Arguments: map[string]any{"name": "key"},
	})
	require.NoError(t, err)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, key!", text.Text)

	// The env key takes precedence, so no key file is generated.
	assert.NoFileExists(t, filepath.Join(server.DataDir, "signing.key"))
}

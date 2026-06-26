package e2e

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"path/filepath"
	"testing"

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

	srv := startServerEnv(t, "SIGNING_KEY="+base64.StdEncoding.EncodeToString(key))

	_, tokens := fullHandshake(t, srv.appURL)
	token := tokens["access_token"].(string)

	cs, err := connect(t, srv.appURL, bearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = cs.Close() })

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "hello_world",
		Arguments: map[string]any{"name": "key"},
	})
	require.NoError(t, err)
	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, key!", text.Text)

	// The env key takes precedence, so no key file is generated.
	assert.NoFileExists(t, filepath.Join(srv.dataDir, "signing.key"))
}

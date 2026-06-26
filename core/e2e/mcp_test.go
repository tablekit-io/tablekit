package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func connect(t *testing.T, appURL string, client *http.Client) (*mcp.ClientSession, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	c := mcp.NewClient(&mcp.Implementation{Name: "e2e", Version: "0"}, nil)
	return c.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:   appURL + "/mcp",
		HTTPClient: client,
	}, nil)
}

func TestMCPListAndCallTool(t *testing.T) {
	srv := startServer(t)
	_, tokens := fullHandshake(t, srv.appURL)
	token := tokens["access_token"].(string)

	cs, err := connect(t, srv.appURL, bearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = cs.Close() })

	ctx := context.Background()

	list, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, list.Tools, 1)
	tool := list.Tools[0]
	assert.Equal(t, "hello_world", tool.Name)
	assert.NotNil(t, tool.OutputSchema)
	require.NotNil(t, tool.Annotations)
	assert.True(t, tool.Annotations.ReadOnlyHint)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "hello_world",
		Arguments: map[string]any{"name": "omran"},
	})
	require.NoError(t, err)
	require.Len(t, res.Content, 1)
	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, omran!", text.Text)
	structured := res.StructuredContent.(map[string]any)
	assert.Equal(t, "Hello, omran!", structured["greeting"])
}

func TestMCPUnauthenticatedRejected(t *testing.T) {
	srv := startServer(t)
	// No bearer token → the MCP handshake must fail.
	_, err := connect(t, srv.appURL, http.DefaultClient)
	assert.Error(t, err)
}

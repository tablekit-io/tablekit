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
	server := startServer(t)
	_, tokens := fullHandshake(t, server.appURL)
	token := tokens["access_token"].(string)

	clientSession, err := connect(t, server.appURL, bearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	ctx := context.Background()

	list, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	// ListTools is the protocol-level listing: it returns every registered tool
	// (hello_world plus the interactive widget and its app-only data loader).
	// App-only visibility is a host-side filter, not a protocol one, so look the
	// tool up by name rather than asserting it's the only one.
	byName := make(map[string]*mcp.Tool, len(list.Tools))
	for _, listed := range list.Tools {
		byName[listed.Name] = listed
	}
	tool := byName["hello_world"]
	require.NotNil(t, tool)
	assert.Equal(t, "hello_world", tool.Name)
	assert.NotNil(t, tool.OutputSchema)
	require.NotNil(t, tool.Annotations)
	assert.True(t, tool.Annotations.ReadOnlyHint)

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "hello_world",
		Arguments: map[string]any{"name": "omran"},
	})
	require.NoError(t, err)
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, omran!", text.Text)
	structured := result.StructuredContent.(map[string]any)
	assert.Equal(t, "Hello, omran!", structured["greeting"])
}

func TestMCPUnauthenticatedRejected(t *testing.T) {
	server := startServer(t)
	// No bearer token → the MCP handshake must fail.
	_, err := connect(t, server.appURL, http.DefaultClient)
	assert.Error(t, err)
}

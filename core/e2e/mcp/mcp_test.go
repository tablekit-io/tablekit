package mcp

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"core/e2e/harness"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPListAndCallTool(t *testing.T) {
	server := harness.StartServer(t)
	_, tokens := harness.FullHandshake(t, server.AppURL)
	token := tokens["access_token"].(string)

	clientSession, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	ctx := context.Background()

	list, err := clientSession.ListTools(ctx, &mcpsdk.ListToolsParams{})
	require.NoError(t, err)
	// ListTools is the protocol-level listing: it returns every registered tool.
	// App-only visibility is a host-side filter, not a protocol one, so look the
	// tool up by name rather than asserting it's the only one.
	byName := make(map[string]*mcpsdk.Tool, len(list.Tools))
	for _, listed := range list.Tools {
		byName[listed.Name] = listed
	}
	tool := byName["list_available_databases"]
	require.NotNil(t, tool)
	assert.Equal(t, "list_available_databases", tool.Name)
	assert.NotNil(t, tool.OutputSchema)
	require.NotNil(t, tool.Annotations)
	assert.True(t, tool.Annotations.ReadOnlyHint)

	result, err := clientSession.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "list_available_databases",
	})
	require.NoError(t, err)
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(*mcpsdk.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "database(s) configured.")
}

func TestMCPUnauthenticatedRejected(t *testing.T) {
	server := harness.StartServer(t)
	// No bearer token → the MCP handshake must fail.
	_, err := harness.Connect(t, server.AppURL, http.DefaultClient)
	assert.Error(t, err)
}

func TestMCPRequiresBearer(t *testing.T) {
	server := harness.StartServer(t)
	response, err := http.Post(server.AppURL+"/mcp", "application/json", strings.NewReader("{}"))
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	assert.Contains(t, response.Header.Get("WWW-Authenticate"), "resource_metadata")
}

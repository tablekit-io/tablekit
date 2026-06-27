package mcp

import (
	"context"
	"testing"

	"core/mcp/ui"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectInMemory wires a client session to newServer() over the SDK's
// in-memory transport — exercises the MCP protocol with no HTTP or auth.
func connectInMemory(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverT, clientT := mcp.NewInMemoryTransports()

	server := newServer()
	ss, err := server.Connect(ctx, serverT, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ss.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	clientSession, err := client.Connect(ctx, clientT, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })
	return clientSession
}

// toolsByName lists the server's tools and indexes them by name.
func toolsByName(t *testing.T, clientSession *mcp.ClientSession) map[string]*mcp.Tool {
	t.Helper()
	result, err := clientSession.ListTools(context.Background(), &mcp.ListToolsParams{})
	require.NoError(t, err)
	byName := make(map[string]*mcp.Tool, len(result.Tools))
	for _, tool := range result.Tools {
		byName[tool.Name] = tool
	}
	return byName
}

func TestListToolsExposesAnnotationsAndSchema(t *testing.T) {
	clientSession := connectInMemory(t)

	tool := toolsByName(t, clientSession)["hello_world"]
	require.NotNil(t, tool)
	assert.NotNil(t, tool.OutputSchema, "tool should advertise an output schema")

	require.NotNil(t, tool.Annotations)
	assert.True(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, tool.Annotations.IdempotentHint)
	require.NotNil(t, tool.Annotations.DestructiveHint)
	assert.False(t, *tool.Annotations.DestructiveHint)
	require.NotNil(t, tool.Annotations.OpenWorldHint)
	assert.False(t, *tool.Annotations.OpenWorldHint)
}

// uiMeta extracts the _meta.ui map a tool advertises, or nil if absent.
func uiMeta(tool *mcp.Tool) map[string]any {
	ui, ok := tool.Meta["ui"].(map[string]any)
	if !ok {
		return nil
	}
	return ui
}

func TestInteractiveToolsRegistered(t *testing.T) {
	clientSession := connectInMemory(t)
	tools := toolsByName(t, clientSession)

	// Build-independent: both tools are always registered.
	require.NotNil(t, tools["hello_world_interactive"])
	data := tools["hello_world_interactive_data"]
	require.NotNil(t, data)

	// The loader is app-only: _meta.ui.visibility advertises ['app'] regardless
	// of whether the widget has been built (it carries no widget link).
	dataUI := uiMeta(data)
	require.NotNil(t, dataUI, "data tool should carry _meta.ui")
	assert.Equal(t, []any{"app"}, dataUI["visibility"])

	// Build-dependent: the model-facing tool links its widget via
	// _meta.ui.resourceUri only when @tablekit/widgets has been built into the
	// embed dir. A fresh checkout (placeholder manifest) carries no link.
	if ui.WidgetURI("hello_world_interactive") != "" {
		meta := uiMeta(tools["hello_world_interactive"])
		require.NotNil(t, meta, "built interactive tool should carry _meta.ui")
		uri, _ := meta["resourceUri"].(string)
		assert.Contains(t, uri, "ui://tablekit/hello_world_interactive-")
	}
}

func TestWidgetResourceIsServed(t *testing.T) {
	// Only meaningful once the widget is built into the embed dir; a fresh
	// checkout ships the placeholder manifest and serves no UI resources.
	resources := ui.Resources()
	if len(resources) == 0 {
		t.Skip("no built widgets in embed dir (run `bun run build` in widgets/)")
	}

	clientSession := connectInMemory(t)
	list, err := clientSession.ListResources(context.Background(), &mcp.ListResourcesParams{})
	require.NoError(t, err)

	var uri string
	for _, r := range list.Resources {
		if r.Name == "hello_world_interactive" {
			uri = r.URI
			assert.Equal(t, "text/html;profile=mcp-app", r.MIMEType)
		}
	}
	require.NotEmpty(t, uri, "widget resource should be listed")

	read, err := clientSession.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uri})
	require.NoError(t, err)
	require.Len(t, read.Contents, 1)
	assert.Contains(t, read.Contents[0].Text, "<html", "resource should serve the widget HTML")
}

func TestCallToolReturnsStructuredContent(t *testing.T) {
	clientSession := connectInMemory(t)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "hello_world",
		Arguments: map[string]any{"name": "omran"},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)

	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, omran!", text.Text)

	structured, ok := result.StructuredContent.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Hello, omran!", structured["greeting"])
}

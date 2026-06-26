package mcpserver

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelloWorld(t *testing.T) {
	tests := []struct {
		name string
		in   helloInput
		want string
	}{
		{"with name", helloInput{Name: "omran"}, "Hello, omran!"},
		{"empty name", helloInput{}, "Hello, world!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, out, err := helloWorld(context.Background(), nil, tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, out.Greeting)
			require.Len(t, res.Content, 1)
			text, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			assert.Equal(t, tt.want, text.Text)
		})
	}
}

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
	cs, err := client.Connect(ctx, clientT, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func TestListToolsExposesAnnotationsAndSchema(t *testing.T) {
	cs := connectInMemory(t)

	res, err := cs.ListTools(context.Background(), &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, res.Tools, 1)

	tool := res.Tools[0]
	assert.Equal(t, "hello_world", tool.Name)
	assert.NotNil(t, tool.OutputSchema, "tool should advertise an output schema")

	require.NotNil(t, tool.Annotations)
	assert.True(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, tool.Annotations.IdempotentHint)
	require.NotNil(t, tool.Annotations.DestructiveHint)
	assert.False(t, *tool.Annotations.DestructiveHint)
	require.NotNil(t, tool.Annotations.OpenWorldHint)
	assert.False(t, *tool.Annotations.OpenWorldHint)
}

func TestCallToolReturnsStructuredContent(t *testing.T) {
	cs := connectInMemory(t)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "hello_world",
		Arguments: map[string]any{"name": "omran"},
	})
	require.NoError(t, err)
	assert.False(t, res.IsError)

	require.Len(t, res.Content, 1)
	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, omran!", text.Text)

	structured, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Hello, omran!", structured["greeting"])
}

package handlers

import (
	"context"

	"core/mcp/ui"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// helloInteractiveWidget is the widget template name (the @tablekit/widgets
// manifest key) and the tool name stem the donut demo is built around.
const helloInteractiveWidget = "hello_world_interactive"

// helloInteractiveInput is the hello_world_interactive tool's argument schema.
type helloInteractiveInput struct{}

// helloInteractiveOutput is a thin discriminator: the widget shares no state
// with the agent, so the tool just names itself. The host renders the linked
// ui:// widget, which fetches its own data over the MCP Apps bridge.
type helloInteractiveOutput struct {
	Tool string `json:"tool" jsonschema:"the tool that produced this result"`
}

// helloInteractive renders the interactive donut widget. The structured result
// is only a discriminator; the real payload is loaded by the widget calling the
// app-only hello_world_interactive_data tool over the bridge.
func helloInteractive(_ context.Context, _ *mcp.CallToolRequest, _ helloInteractiveInput) (*mcp.CallToolResult, helloInteractiveOutput, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: "Rendering an interactive donut chart with random example data.",
		}},
	}, helloInteractiveOutput{Tool: helloInteractiveWidget}, nil
}

// registerHelloWorldInteractive adds the model-facing tool that renders the
// donut widget. The _meta.ui.resourceUri links the widget the host prefetches
// and renders; it's resolved from the build manifest (empty until widgets are
// built, in which case the tool simply advertises no widget).
func registerHelloWorldInteractive(s *mcp.Server) {
	interactive := &mcp.Tool{
		Name:        "hello_world_interactive",
		Description: "Renders an interactive donut chart with random example data, using MCP Apps.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			DestructiveHint: pointer(false),
			OpenWorldHint:   pointer(false),
		},
	}
	if uri := ui.WidgetURI(helloInteractiveWidget); uri != "" {
		interactive.Meta = mcp.Meta{"ui": map[string]any{"resourceUri": uri}}
	}
	mcp.AddTool(s, interactive, helloInteractive)
}

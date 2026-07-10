package shared

import (
	"core/mcp/ui"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// chartWidget is the @tablekit/widgets template both render tools link. The
// widget reads the render tool's arguments (the axis/series mapping) from the
// host, discriminates on the tool name in the result, and calls fetch_chart_data
// over the bridge to pull the rows it renders.
const ChartWidget = "chart_renderer"

// ChartHint nudges the agent toward the dedicated chart tools whenever it has
// rows in hand (query_database with include_results, or read_results). Surfaced
// both as a structured hints_for_ai_agents field and appended to the text content
// so it is seen regardless of how the client reads the result. The goal is
// consistent tablekit visualizations users can build muscle memory around,
// rather than the agent hand-formatting charts itself.
const ChartHint = "If the user wants a visualized chart, prefer the show_bar_line_area_chart tool or show_pie_donut_sunburst_chart tool (pass this result_key) instead of drawing charts yourself — this keeps tablekit's visualizations consistent so users benefit from muscle memory."

// ColumnInfo is one result column in a tool's structured output.
type ColumnInfo struct {
	Name string `json:"name" jsonschema:"the column name"`
}

// ChartRenderOutput is the thin discriminator the render tools return. The widget
// shares no state with the agent: it reads the render tool's arguments from the
// host and loads rows via fetch_chart_data, so the tool only names itself.
type ChartRenderOutput struct {
	Tool             string   `json:"tool" jsonschema:"the tool that produced this result"`
	HintsForAIAgents []string `json:"hints_for_ai_agents,omitempty" jsonschema:"guidance for the calling AI agent on how best to use this result"`
}

// ChartWidgetMeta is the tool _meta that links a render tool to the chart widget
// template, or nil when the widget isn't built yet. It advertises the same ui://
// URI under two keys so both host families pick it up: `ui.resourceUri` is the
// MCP Apps convention (Claude and other MCP-UI hosts), while ChatGPT's Apps SDK
// binds the tool call to its template via `openai/outputTemplate` — without that
// key ChatGPT renders the shell but never populates window.openai.toolInput /
// toolOutput inside the iframe, so the widget can't learn its query_key.
func ChartWidgetMeta() mcp.Meta {
	uri := ui.WidgetURI(ChartWidget)
	if uri == "" {
		return nil
	}
	return mcp.Meta{
		"ui":                    map[string]any{"resourceUri": uri},
		"openai/outputTemplate": uri,
	}
}

// WidgetBridgeMeta is the _meta for an app-only tool the chart widget calls over
// the host bridge (fetch_chart_data, get_export_url). `ui.visibility=["app"]`
// hides it from the model while keeping it callable from the UI (the MCP Apps
// convention); ChatGPT additionally requires `openai/widgetAccessible=true` to
// permit component-initiated tool calls, without which window.openai.callTool
// rejects the request and the widget can never load its rows.
func WidgetBridgeMeta() mcp.Meta {
	return mcp.Meta{
		"ui":                      map[string]any{"visibility": []string{"app"}},
		"openai/widgetAccessible": true,
	}
}

// ChartRenderResult builds the discriminator result a render tool returns. tool
// is the tool name the widget branches on; label is the human phrase for the
// summary line.
func ChartRenderResult(tool, label string) (*mcp.CallToolResult, ChartRenderOutput, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Rendering a " + label + " from the stored query."}},
	}, ChartRenderOutput{Tool: tool}, nil
}

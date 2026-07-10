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

// ChartRenderOutput is the discriminator the render tools return, plus a copy of
// the render arguments the widget needs (query_key + the axis/series mapping).
//
// The widget could read those arguments from the host's tool-input notification,
// but ChatGPT mis-binds that: when the agent runs query_database (with inline
// results) and then a render tool in one turn, ChatGPT delivers query_database's
// arguments as the widget's tool-input — so query_key never arrives. The
// tool-RESULT, by contrast, is delivered against the right call. Echoing the
// arguments here lets the widget read them from the result and stay correct
// regardless of how the host routes tool-input. Args is the render tool's own
// input struct (query_key, x/y or value/layers, formatting), marshaled as-is.
type ChartRenderOutput struct {
	Tool             string   `json:"tool" jsonschema:"the tool that produced this result"`
	Args             any      `json:"args,omitempty" jsonschema:"the render arguments (query_key and axis/series mapping) the widget renders from"`
	HintsForAIAgents []string `json:"hints_for_ai_agents,omitempty" jsonschema:"guidance for the calling AI agent on how best to use this result"`
}

// ChartWidgetMeta is the tool _meta that links a render tool to the chart widget
// template via the MCP Apps `_meta.ui.resourceUri` convention, or nil when the
// widget isn't built yet. The host prefetches that ui:// resource and renders it
// in a sandboxed iframe; the widget then loads its rows over the postMessage
// bridge. This is the whole contract for MCP-UI hosts — Claude and ChatGPT alike
// (ChatGPT speaks the same postMessage protocol here, so no openai/* keys are
// needed; adding openai/outputTemplate would flip it into Apps-SDK mode and it
// would stop delivering the tool-input the widget needs).
func ChartWidgetMeta() mcp.Meta {
	uri := ui.WidgetURI(ChartWidget)
	if uri == "" {
		return nil
	}
	return mcp.Meta{"ui": map[string]any{"resourceUri": uri}}
}

// WidgetBridgeMeta is the _meta for an app-only tool the chart widget calls over
// the host bridge (fetch_chart_data, get_export_url). `ui.visibility=["app"]`
// hides it from the model while keeping it callable from the UI — the MCP Apps
// gate the host honours for component-initiated tools/call.
func WidgetBridgeMeta() mcp.Meta {
	return mcp.Meta{"ui": map[string]any{"visibility": []string{"app"}}}
}

// ChartRenderResult builds the discriminator result a render tool returns. tool
// is the tool name the widget branches on; label is the human phrase for the
// summary line; args is the render tool's own input, echoed into the result so
// the widget can read query_key + the mapping from the result rather than the
// host's (sometimes mis-bound) tool-input notification.
func ChartRenderResult(tool, label string, args any) (*mcp.CallToolResult, ChartRenderOutput, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Rendering a " + label + " from the stored query."}},
	}, ChartRenderOutput{Tool: tool, Args: args}, nil
}

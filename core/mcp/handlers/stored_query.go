package handlers

import (
	"context"
	"fmt"

	"core/engine"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// chartRenderOutput is the thin discriminator the render tools return. The widget
// shares no state with the agent: it reads the render tool's arguments from the
// host and loads rows via fetch_chart_data, so the tool only names itself.
type chartRenderOutput struct {
	Tool string `json:"tool" jsonschema:"the tool that produced this result"`
}

// requireQuery confirms a stored query exists for key, turning an unknown key
// into a user-facing error.
func (h *Handlers) requireQuery(ctx context.Context, key string) error {
	descriptor, err := h.Queries.Get(ctx, key)
	if err != nil {
		return err
	}
	if descriptor == nil {
		return fmt.Errorf("unknown query_key %q (run run_query first)", key)
	}
	return nil
}

// chartRenderResult builds the discriminator result a render tool returns. tool
// is the tool name the widget branches on; label is the human phrase for the
// summary line.
func chartRenderResult(tool, label string) (*mcp.CallToolResult, chartRenderOutput, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Rendering a " + label + " from the stored query."}},
	}, chartRenderOutput{Tool: tool}, nil
}

// Chart/export sizing: the render and export paths fetch the whole result, so
// they raise the row and byte caps well above the run_query page size.
const (
	chartMaxRows  = 100_000
	chartMaxBytes = 16 << 20 // 16 MiB
)

// chartHint nudges the agent toward the dedicated chart tools whenever it has
// rows in hand (run_query with include_results, or retrieve_results). Surfaced
// both as a structured hint_for_agents field and appended to the text content so
// it is seen regardless of how the client reads the result. The goal is
// consistent tablekit visualizations users can build muscle memory around,
// rather than the agent hand-formatting charts itself.
const chartHint = "If the user wants a visualization, prefer the render_cartesian_series_chart or render_proportional_chart tools (pass this result_key) instead of formatting the data yourself — this keeps tablekit's visualizations consistent so users benefit from muscle memory."

// enginePage builds engine.PageOptions; a zero maxBytes lets the engine apply
// its default.
func enginePage(offset, limit, maxBytes int) engine.PageOptions {
	return engine.PageOptions{Offset: offset, Limit: limit, MaxBytes: maxBytes}
}

// toColumnInfos wraps column names as the structured columnInfo list tools return.
func toColumnInfos(columns []string) []columnInfo {
	infos := make([]columnInfo, len(columns))
	for i, name := range columns {
		infos[i] = columnInfo{Name: name}
	}
	return infos
}

// moreSuffix renders a short " (more rows available)" note for tool summaries.
func moreSuffix(hasMore bool) string {
	if hasMore {
		return " (more rows available)"
	}
	return ""
}

// projectColumns narrows a result to the requested columns, preserving the
// requested order and silently dropping names the result doesn't have. An empty
// request returns the columns and rows unchanged. It does not mutate the input.
func projectColumns(columns []string, rows []map[string]any, requested []string) (projectedColumns []string, projectedRows []map[string]any) {
	if len(requested) == 0 {
		return columns, rows
	}
	present := make(map[string]bool, len(columns))
	for _, name := range columns {
		present[name] = true
	}
	kept := make([]string, 0, len(requested))
	for _, name := range requested {
		if present[name] {
			kept = append(kept, name)
		}
	}
	out := make([]map[string]any, len(rows))
	for i, row := range rows {
		narrowed := make(map[string]any, len(kept))
		for _, name := range kept {
			if value, ok := row[name]; ok {
				narrowed[name] = value
			}
		}
		out[i] = narrowed
	}
	return kept, out
}

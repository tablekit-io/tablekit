package handlers

import (
	"context"
	"fmt"

	"core/helpers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// fetchChartDataInput is the fetch_chart_data loader's argument schema.
type fetchChartDataInput struct {
	QueryKey string `json:"query_key" jsonschema:"the result_key returned by run_query"`
}

// fetchChartDataOutput is the full result a chart widget renders: every row (up
// to the chart cap), the columns, and the stored SQL for display.
type fetchChartDataOutput struct {
	Columns  []string         `json:"columns" jsonschema:"the result column names, in order"`
	Rows     []map[string]any `json:"rows" jsonschema:"all result rows (up to the chart row cap)"`
	RowCount int              `json:"row_count" jsonschema:"number of rows returned"`
	SQL      string           `json:"sql" jsonschema:"the stored SQL that produced these rows"`
}

// fetchChartData loads the whole result of a stored query for a chart widget. It
// re-runs the stored SQL with the raised chart row/byte caps and no offset.
// App-only: called by the chart widget over the MCP Apps bridge, hidden from the
// agent.
func (h *Handlers) fetchChartData(ctx context.Context, _ *mcp.CallToolRequest, in fetchChartDataInput) (*mcp.CallToolResult, fetchChartDataOutput, error) {
	descriptor, err := h.Queries.Get(ctx, in.QueryKey)
	if err != nil {
		return nil, fetchChartDataOutput{}, err
	}
	if descriptor == nil {
		return nil, fetchChartDataOutput{}, fmt.Errorf("unknown query_key %q (run run_query first)", in.QueryKey)
	}

	result, _, err := h.Engine.RunReadOnlyPage(ctx, descriptor.Database, descriptor.SQL, enginePage(0, chartMaxRows, chartMaxBytes))
	if err != nil {
		return nil, fetchChartDataOutput{}, err
	}
	if result.Truncated {
		return nil, fetchChartDataOutput{}, fmt.Errorf("result is too large to chart (exceeds the chart size cap)")
	}

	out := fetchChartDataOutput{
		Columns:  result.Columns,
		Rows:     result.Rows,
		RowCount: result.RowCount,
		SQL:      descriptor.SQL,
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Loaded %d row(s) for charting.", result.RowCount)}},
	}, out, nil
}

// registerFetchChartData adds the app-only chart data loader. The
// _meta.ui.visibility=['app'] marks it app-only — the host hides it from the
// model and only honours it when a chart widget calls it over the bridge.
func (h *Handlers) registerFetchChartData(s *mcp.Server) {
	dataTool := &mcp.Tool{
		Name:        "fetch_chart_data",
		Description: "Returns the full result of a stored query for a chart widget to render. App-only: called by the chart widgets over the MCP Apps bridge, hidden from the agent.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(true),
		},
	}
	dataTool.Meta = mcp.Meta{"ui": map[string]any{"visibility": []string{"app"}}}
	mcp.AddTool(s, dataTool, h.fetchChartData)
}

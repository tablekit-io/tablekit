// Package fetchchartdata implements the app-only fetch_chart_data MCP tool.
package fetchchartdata

import (
	"context"
	_ "embed"
	"fmt"

	"core/helpers"
	"core/mcp/handlers/shared"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed schema.json
var schemaJSON []byte

// input is the fetch_chart_data loader's argument schema.
type input struct {
	QueryKey string `json:"query_key"`
}

// output is the full result a chart widget renders: every row (up to the chart
// cap), the columns, and the stored SQL for display.
type output struct {
	Columns  []string         `json:"columns" jsonschema:"the result column names, in order"`
	Rows     []map[string]any `json:"rows" jsonschema:"all result rows (up to the chart row cap)"`
	RowCount int              `json:"row_count" jsonschema:"number of rows returned"`
	SQL      string           `json:"sql" jsonschema:"the stored SQL that produced these rows"`
}

// Register adds the app-only chart data loader. The _meta.ui.visibility=['app']
// marks it app-only — the host hides it from the model and only honours it when
// a chart widget calls it over the bridge.
func Register(s *mcp.Server, deps shared.Deps) {
	tool := &mcp.Tool{
		Name:        "fetch_chart_data",
		Description: "Returns the full result of a stored query for a chart widget to render. App-only: called by the chart widgets over the MCP Apps bridge, hidden from the agent.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(true),
		},
	}
	tool.Meta = mcp.Meta{"ui": map[string]any{"visibility": []string{"app"}}}
	tool.InputSchema = shared.InputSchema[input](schemaJSON)
	mcp.AddTool(s, tool, handle(deps))
}

// handle loads the whole result of a stored query for a chart widget. It re-runs
// the stored SQL with the raised chart row/byte caps and no offset. App-only:
// called by the chart widget over the MCP Apps bridge, hidden from the agent.
func handle(deps shared.Deps) mcp.ToolHandlerFor[input, output] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in input) (*mcp.CallToolResult, output, error) {
		descriptor, err := deps.Queries.Get(ctx, in.QueryKey)
		if err != nil {
			return nil, output{}, err
		}
		if descriptor == nil {
			return nil, output{}, fmt.Errorf("unknown query_key %q (run query_database first)", in.QueryKey)
		}

		result, _, err := deps.Engine.RunReadOnlyPage(ctx, descriptor.Database, descriptor.SQL, shared.EnginePage(0, shared.ChartMaxRows, shared.ChartMaxBytes))
		if err != nil {
			return nil, output{}, err
		}
		if result.Truncated {
			return nil, output{}, fmt.Errorf("result is too large to chart (exceeds the chart size cap)")
		}

		out := output{
			Columns:  result.Columns,
			Rows:     result.Rows,
			RowCount: result.RowCount,
			SQL:      descriptor.SQL,
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Loaded %d row(s) for charting.", result.RowCount)}},
		}, out, nil
	}
}

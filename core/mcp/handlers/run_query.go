package handlers

import (
	"context"
	"fmt"

	"core/helpers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Pagination sizing shared by the stored-query tools. defaultLimit is the window
// run_query previews and retrieve_results uses when no limit is given; maxLimit
// caps a single retrieve_results window.
const (
	defaultLimit = 128
	maxLimit     = 2048
)

// columnInfo is one result column in a tool's structured output.
type columnInfo struct {
	Name string `json:"name" jsonschema:"the column name"`
}

// runQueryInput is the run_query tool's argument schema.
type runQueryInput struct {
	Database       string `json:"database" jsonschema:"name of the configured database to query (see list_databases)"`
	SQL            string `json:"sql" jsonschema:"the read-only SQL (SELECT / WITH ... SELECT) to execute; no writes"`
	Description    string `json:"description" jsonschema:"required plain-language description of what this query is for"`
	IncludeResults bool   `json:"include_results,omitempty" jsonschema:"when true, inline the first page of rows in the result (only do this for small results)"`
}

// runQueryOutput is run_query's structured result: stats plus the result_key the
// other tools take. Rows are inlined only when include_results was set.
type runQueryOutput struct {
	ResultKey    string           `json:"result_key" jsonschema:"the key identifying this stored query; pass it to retrieve_results, render_*_chart and get_export_url"`
	RowCount     int              `json:"row_count" jsonschema:"number of rows in this first page (at most default_limit)"`
	HasMore      bool             `json:"has_more" jsonschema:"true when more rows exist beyond this first page"`
	DefaultLimit int              `json:"default_limit" jsonschema:"the page size retrieve_results uses by default"`
	Columns      []columnInfo     `json:"columns" jsonschema:"the result columns, in order"`
	Rows         []map[string]any `json:"rows,omitempty" jsonschema:"the first page of rows, only present when include_results was set"`
}

// runQuery executes read-only SQL, stores the query descriptor, and returns a
// result_key the agent passes to retrieve_results / render_*_chart / get_export_url.
// Unlike run_sql it persists nothing but the query text (not the rows): the other
// tools re-run the stored SQL against the live database.
func (h *Handlers) runQuery(ctx context.Context, _ *mcp.CallToolRequest, in runQueryInput) (*mcp.CallToolResult, runQueryOutput, error) {
	if in.Description == "" {
		return nil, runQueryOutput{}, fmt.Errorf("description is required")
	}

	result, hasMore, err := h.Engine.RunReadOnlyPage(ctx, in.Database, in.SQL, enginePage(0, defaultLimit, 0))
	if err != nil {
		return nil, runQueryOutput{}, err
	}

	key, err := h.Queries.Save(ctx, in.Database, in.SQL, in.Description)
	if err != nil {
		return nil, runQueryOutput{}, err
	}

	out := runQueryOutput{
		ResultKey:    key,
		RowCount:     result.RowCount,
		HasMore:      hasMore,
		DefaultLimit: defaultLimit,
		Columns:      toColumnInfos(result.Columns),
	}
	if in.IncludeResults {
		out.Rows = result.Rows
	}

	summary := fmt.Sprintf(
		"Stored query %s against %q: %d row(s) in the first page%s. "+
			"Pass result_key to retrieve_results (to page), render_cartesian_series_chart / render_proportional_chart (to chart), or get_export_url (to download).",
		key, in.Database, result.RowCount, moreSuffix(hasMore),
	)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, out, nil
}

// registerRunQuery adds the run_query tool. Read-only, but touches external
// databases, so OpenWorldHint is true.
func (h *Handlers) registerRunQuery(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "run_query",
		Description: "Runs read-only SQL against a configured database, stores the query, and returns a result_key plus first-page stats. Use the key with retrieve_results to paginate, render_cartesian_series_chart / render_proportional_chart to visualize, or get_export_url to download CSV/JSON. Rows are not stored — each follow-up re-runs the query against live data. Use list_databases to discover database names.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(true),
		},
	}, h.runQuery)
}

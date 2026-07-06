package handlers

import (
	"context"
	"fmt"

	"core/helpers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Pagination sizing shared by the stored-query tools. defaultLimit is the window
// query_database previews and read_results uses when no limit is given; maxLimit
// caps a single read_results window.
const (
	defaultLimit = 128
	maxLimit     = 2048
)

// columnInfo is one result column in a tool's structured output.
type columnInfo struct {
	Name string `json:"name" jsonschema:"the column name"`
}

// queryDatabaseInput is the query_database tool's argument schema.
type queryDatabaseInput struct {
	Database       string `json:"database" jsonschema:"name of the database to query (see list_available_databases)"`
	SQL            string `json:"sql" jsonschema:"the read-only SQL (SELECT / WITH ... SELECT) to execute; no writes. Annotate it with concise -- comments explaining what each ambiguous part does and why, since the stored SQL is read by human reviewers"`
	Description    string `json:"description" jsonschema:"required, concise plain-language description of what this query is about and why it is being run"`
	IncludeResults bool   `json:"include_results,omitempty" jsonschema:"when true, inline the first page of rows in the result (only do this for small result sets)"`
}

// queryDatabaseOutput is query_database's structured result: stats plus the result_key
// the other tools take. Rows are inlined only when include_results was set.
type queryDatabaseOutput struct {
	ResultKey        string           `json:"result_key" jsonschema:"the key identifying this stored query; pass it to read_results, show_bar_line_area_chart or show_pie_donut_sunburst_chart"`
	RowCount         int              `json:"row_count" jsonschema:"number of rows in this first page (at most default_limit)"`
	HasMore          bool             `json:"has_more" jsonschema:"true when more rows exist beyond this first page"`
	DefaultLimit     int              `json:"default_limit" jsonschema:"the page size read_results uses by default"`
	Columns          []columnInfo     `json:"columns" jsonschema:"the result columns, in order"`
	Rows             []map[string]any `json:"rows,omitempty" jsonschema:"the first page of rows, only present when include_results was set"`
	HintsForAIAgents []string         `json:"hints_for_ai_agents,omitempty" jsonschema:"guidance for the calling AI agent on how best to use this result"`
}

// queryDatabase executes read-only SQL, stores the query descriptor, and returns a
// result_key the agent passes to read_results / show_bar_line_area_chart /
// show_pie_donut_sunburst_chart / get_export_url. It persists nothing but the query
// text (not the rows): the other tools re-run the stored SQL against the live database.
func (h *Handlers) queryDatabase(ctx context.Context, _ *mcp.CallToolRequest, in queryDatabaseInput) (*mcp.CallToolResult, queryDatabaseOutput, error) {
	if in.Description == "" {
		return nil, queryDatabaseOutput{}, fmt.Errorf("description is required")
	}

	result, hasMore, err := h.Engine.RunReadOnlyPage(ctx, in.Database, in.SQL, enginePage(0, defaultLimit, 0))
	if err != nil {
		return nil, queryDatabaseOutput{}, err
	}

	key, err := h.Queries.Save(ctx, in.Database, in.SQL, in.Description)
	if err != nil {
		return nil, queryDatabaseOutput{}, err
	}

	out := queryDatabaseOutput{
		ResultKey:    key,
		RowCount:     result.RowCount,
		HasMore:      hasMore,
		DefaultLimit: defaultLimit,
		Columns:      toColumnInfos(result.Columns),
	}
	summary := fmt.Sprintf(
		"Stored query %s against %q: %d row(s) in the first page%s. "+
			"Pass result_key to read_results (to paginate over the result set), show_bar_line_area_chart / show_pie_donut_sunburst_chart to display charts for the user.",
		key, in.Database, result.RowCount, moreSuffix(hasMore),
	)
	// When rows are inlined the agent has data in hand, so nudge it toward the
	// chart tools (same as read_results).
	if in.IncludeResults {
		out.Rows = result.Rows
		out.HintsForAIAgents = []string{chartHint}
		summary += " " + chartHint
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, out, nil
}

// registerQueryDatabase adds the query_database tool. Read-only, but touches external
// databases, so OpenWorldHint is true.
func (h *Handlers) registerQueryDatabase(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "query_database",
		Description: "Run read-only SQL on a database. It stores the query, and returns a result_key plus first-page stats (in order to be token efficient). Use the result key with read_results to paginate over the results. Use the result key with show_bar_line_area_chart / show_pie_donut_sunburst_chart to visualize with charts. Rows are not stored — each follow-up re-runs the query against live data. Use list_available_databases to discover database names. The stored SQL is reviewed by humans, so always annotate it with concise -- comments explaining what each ambiguous part of the query does and why it is being done.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(true),
		},
	}, h.queryDatabase)
}

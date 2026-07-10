// Package querydatabase implements the query_database MCP tool.
package querydatabase

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

// input is the query_database tool's argument schema. Descriptions live in
// schema.json; the struct only decodes.
type input struct {
	Database       string `json:"database"`
	SQL            string `json:"sql"`
	Description    string `json:"description"`
	IncludeResults bool   `json:"include_results,omitempty"`
}

// output is query_database's structured result: stats plus the result_key the
// other tools take. Rows are inlined only when include_results was set.
type output struct {
	ResultKey        string              `json:"result_key" jsonschema:"the key identifying this stored query; pass it to read_results, show_bar_line_area_chart or show_pie_donut_sunburst_chart"`
	RowCount         int                 `json:"row_count" jsonschema:"number of rows in this first page (at most default_limit)"`
	HasMore          bool                `json:"has_more" jsonschema:"true when more rows exist beyond this first page"`
	DefaultLimit     int                 `json:"default_limit" jsonschema:"the page size read_results uses by default"`
	Columns          []shared.ColumnInfo `json:"columns" jsonschema:"the result columns, in order"`
	Rows             []map[string]any    `json:"rows,omitempty" jsonschema:"the first page of rows, only present when include_results was set"`
	HintsForAIAgents []string            `json:"hints_for_ai_agents,omitempty" jsonschema:"guidance for the calling AI agent on how best to use this result"`
}

// Register adds the query_database tool. Read-only, but touches external
// databases, so OpenWorldHint is true.
func Register(s *mcp.Server, deps shared.Deps) {
	tool := &mcp.Tool{
		Name:        "query_database",
		Description: "Run read-only SQL on a database. It stores the query, and returns a result_key plus first-page stats (in order to be token efficient). Use the result key with read_results to paginate over the results. Use the result key with show_bar_line_area_chart / show_pie_donut_sunburst_chart to visualize with charts. Rows are not stored — each follow-up re-runs the query against live data. Use list_available_databases to discover database names. The stored SQL is reviewed by humans, so always annotate it with concise -- comments explaining what each ambiguous part of the query does and why it is being done.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(true),
		},
	}
	tool.InputSchema = shared.InputSchema[input](schemaJSON)
	mcp.AddTool(s, tool, handle(deps))
}

// handle executes read-only SQL, stores the query descriptor, and returns a
// result_key the agent passes to read_results / show_bar_line_area_chart /
// show_pie_donut_sunburst_chart / get_export_url. It persists nothing but the query
// text (not the rows): the other tools re-run the stored SQL against the live database.
func handle(deps shared.Deps) mcp.ToolHandlerFor[input, output] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in input) (*mcp.CallToolResult, output, error) {
		if in.Description == "" {
			return nil, output{}, fmt.Errorf("description is required")
		}

		// Attribute the stored query to the calling client (queries.client_id).
		clientID, err := shared.ClientID(req)
		if err != nil {
			return nil, output{}, err
		}

		// Pin the physical database on first query: derive its stable identity and
		// mint/match a database_id. A repointed name is caught here and on re-run.
		databaseID, err := deps.Databases.Resolve(ctx, in.Database)
		if err != nil {
			return nil, output{}, err
		}

		result, hasMore, err := deps.Engine.RunReadOnlyPage(ctx, in.Database, in.SQL, shared.EnginePage(0, shared.DefaultLimit, 0))
		if err != nil {
			return nil, output{}, err
		}

		key, err := deps.Queries.Save(ctx, databaseID, clientID, in.SQL, in.Description)
		if err != nil {
			return nil, output{}, err
		}

		out := output{
			ResultKey:    key.String(),
			RowCount:     result.RowCount,
			HasMore:      hasMore,
			DefaultLimit: shared.DefaultLimit,
			Columns:      shared.ToColumnInfos(result.Columns),
		}
		summary := fmt.Sprintf(
			"Stored query %s against %q: %d row(s) in the first page%s. "+
				"Pass result_key to read_results (to paginate over the result set), show_bar_line_area_chart / show_pie_donut_sunburst_chart to display charts for the user.",
			key, in.Database, result.RowCount, shared.MoreSuffix(hasMore),
		)
		// When rows are inlined the agent has data in hand, so nudge it toward the
		// chart tools (same as read_results).
		if in.IncludeResults {
			out.Rows = result.Rows
			out.HintsForAIAgents = []string{shared.ChartHint}
			summary += " " + shared.ChartHint
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: summary}},
		}, out, nil
	}
}

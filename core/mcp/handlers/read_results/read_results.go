// Package readresults implements the read_results MCP tool.
package readresults

import (
	"context"
	_ "embed"
	"fmt"

	"core/helpers"
	"core/mcp/handlers/shared"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

//go:embed schema.json
var schemaJSON []byte

// input is the read_results tool's argument schema. Descriptions live in
// schema.json; the struct only decodes.
type input struct {
	Key     string   `json:"key"`
	Skip    int      `json:"skip,omitempty"`
	Limit   int      `json:"limit,omitempty"`
	Columns []string `json:"columns,omitempty"`
}

// output is one paginated window of a stored query's rows.
type output struct {
	Skip             int              `json:"skip" jsonschema:"the offset this window started at"`
	Limit            int              `json:"limit" jsonschema:"the page size used for this window"`
	Columns          []string         `json:"columns" jsonschema:"the returned column names, in order (after any subset filtering)"`
	Rows             []map[string]any `json:"rows" jsonschema:"the rows in this window"`
	RowsReturned     int              `json:"rows_returned" jsonschema:"number of rows in this window"`
	HasMore          bool             `json:"has_more" jsonschema:"true when more rows exist beyond this window"`
	NextSkip         *int             `json:"next_skip" jsonschema:"the skip to pass for the next window, or null when this is the last window"`
	HintsForAIAgents []string         `json:"hints_for_ai_agents,omitempty" jsonschema:"guidance for the calling AI agent on how best to use this result"`
}

// Register adds the read_results tool. Read-only, re-runs the stored query
// against an external database, so OpenWorldHint is true.
func Register(s *mcp.Server, deps shared.Deps) {
	tool := &mcp.Tool{
		Name:        "read_results",
		Description: "Returns a paginated view of a query result. Pass the result_key from query_database, plus optional skip/limit and an optional column subset. Each call re-runs the stored SQL at the requested offset against live data; use has_more / next_skip to page.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(true),
		},
	}
	tool.InputSchema = shared.InputSchema[input](schemaJSON)
	mcp.AddTool(s, tool, handle(deps))
}

// handle pages through a stored query's rows. It re-runs the stored SQL at the
// requested offset/limit (rows are never cached), then optionally narrows to a
// requested subset of columns.
func handle(deps shared.Deps) mcp.ToolHandlerFor[input, output] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in input) (*mcp.CallToolResult, output, error) {
		id, err := uuid.Parse(in.Key)
		if err != nil {
			log.Warn().Str("key", in.Key).Msg("read_results unknown key")
			return nil, output{}, fmt.Errorf("unknown result_key %q (run query_database first)", in.Key)
		}
		descriptor, err := deps.Queries.Get(ctx, id)
		if err != nil {
			log.Error().Str("key", in.Key).Err(err).Msg("read_results descriptor load failed")
			return nil, output{}, err
		}
		if descriptor == nil {
			log.Warn().Str("key", in.Key).Msg("read_results unknown key")
			return nil, output{}, fmt.Errorf("unknown result_key %q (run query_database first)", in.Key)
		}

		skip := max(in.Skip, 0)
		limit := in.Limit
		if limit <= 0 {
			limit = shared.DefaultLimit
		}
		if limit > shared.MaxLimit {
			limit = shared.MaxLimit
		}

		// Confirm the descriptor's database_id still resolves to the same physical
		// database the query was saved against, then run against that name.
		name, err := deps.Databases.Verify(ctx, descriptor.DatabaseID)
		if err != nil {
			log.Warn().Str("key", in.Key).Err(err).Msg("read_results verify failed (repointed?)")
			return nil, output{}, err
		}

		result, hasMore, err := deps.Engine.RunReadOnlyPage(ctx, name, descriptor.SQL, shared.EnginePage(skip, limit, 0))
		if err != nil {
			log.Error().Str("key", in.Key).Err(err).Msg("read_results engine run failed")
			return nil, output{}, err
		}

		columns, rows := shared.ProjectColumns(result.Columns, result.Rows, in.Columns)

		var nextSkip *int
		if hasMore {
			next := skip + limit
			nextSkip = &next
		}

		out := output{
			Skip:             skip,
			Limit:            limit,
			Columns:          columns,
			Rows:             rows,
			RowsReturned:     len(rows),
			HasMore:          hasMore,
			NextSkip:         nextSkip,
			HintsForAIAgents: []string{shared.ChartHint},
		}
		summary := fmt.Sprintf("Rows %d–%d of stored query %s: %d row(s)%s. %s",
			skip, skip+len(rows), in.Key, len(rows), shared.MoreSuffix(hasMore), shared.ChartHint)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: summary}},
		}, out, nil
	}
}

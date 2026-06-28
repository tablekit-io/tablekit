package handlers

import (
	"context"
	"fmt"

	"core/helpers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// retrieveResultsInput is the retrieve_results tool's argument schema.
type retrieveResultsInput struct {
	Key     string   `json:"key" jsonschema:"the result_key returned by run_query"`
	Skip    int      `json:"skip,omitempty" jsonschema:"number of leading rows to skip (OFFSET); defaults to 0"`
	Limit   int      `json:"limit,omitempty" jsonschema:"maximum rows in this window; defaults to 128, capped at 2048"`
	Columns []string `json:"columns,omitempty" jsonschema:"optional subset of column names to return, in the given order"`
}

// retrieveResultsOutput is one paginated window of a stored query's rows.
type retrieveResultsOutput struct {
	Skip          int              `json:"skip" jsonschema:"the offset this window started at"`
	Limit         int              `json:"limit" jsonschema:"the page size used for this window"`
	Columns       []string         `json:"columns" jsonschema:"the returned column names, in order (after any subset filtering)"`
	Rows          []map[string]any `json:"rows" jsonschema:"the rows in this window"`
	RowsReturned  int              `json:"rows_returned" jsonschema:"number of rows in this window"`
	HasMore       bool             `json:"has_more" jsonschema:"true when more rows exist beyond this window"`
	NextSkip      *int             `json:"next_skip" jsonschema:"the skip to pass for the next window, or null when this is the last window"`
	HintForAgents string           `json:"hint_for_agents,omitempty" jsonschema:"guidance for the calling agent on how best to use this result"`
}

// retrieveResults pages through a stored query's rows. It re-runs the stored SQL
// at the requested offset/limit (rows are never cached), then optionally narrows
// to a requested subset of columns.
func (h *Handlers) retrieveResults(ctx context.Context, _ *mcp.CallToolRequest, in retrieveResultsInput) (*mcp.CallToolResult, retrieveResultsOutput, error) {
	descriptor, err := h.Queries.Get(ctx, in.Key)
	if err != nil {
		return nil, retrieveResultsOutput{}, err
	}
	if descriptor == nil {
		return nil, retrieveResultsOutput{}, fmt.Errorf("unknown result_key %q (run run_query first)", in.Key)
	}

	skip := max(in.Skip, 0)
	limit := in.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	result, hasMore, err := h.Engine.RunReadOnlyPage(ctx, descriptor.Database, descriptor.SQL, enginePage(skip, limit, 0))
	if err != nil {
		return nil, retrieveResultsOutput{}, err
	}

	columns, rows := projectColumns(result.Columns, result.Rows, in.Columns)

	var nextSkip *int
	if hasMore {
		next := skip + limit
		nextSkip = &next
	}

	out := retrieveResultsOutput{
		Skip:          skip,
		Limit:         limit,
		Columns:       columns,
		Rows:          rows,
		RowsReturned:  len(rows),
		HasMore:       hasMore,
		NextSkip:      nextSkip,
		HintForAgents: chartHint,
	}
	summary := fmt.Sprintf("Rows %d–%d of stored query %s: %d row(s)%s. %s",
		skip, skip+len(rows), in.Key, len(rows), moreSuffix(hasMore), chartHint)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, out, nil
}

// registerRetrieveResults adds the retrieve_results tool. Read-only, re-runs the
// stored query against an external database, so OpenWorldHint is true.
func (h *Handlers) registerRetrieveResults(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "retrieve_results",
		Description: "Returns a paginated window of a stored query's rows. Pass the result_key from run_query, plus optional skip/limit and an optional column subset. Each call re-runs the stored SQL at the requested offset against live data; use has_more / next_skip to page.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(true),
		},
	}, h.retrieveResults)
}

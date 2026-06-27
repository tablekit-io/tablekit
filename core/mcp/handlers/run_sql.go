package handlers

import (
	"context"
	"fmt"

	"core/engine"
	"core/helpers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// runSQLInput is the run_sql tool's argument schema.
type runSQLInput struct {
	Database string `json:"database" jsonschema:"name of the configured database to query (see list_databases)"`
	SQL      string `json:"sql" jsonschema:"the read-only SQL to execute"`
}

// runSQLOutput mirrors engine.Result: columns, rows as objects, and flags for
// truncation and omitted columns.
type runSQLOutput struct {
	Columns   []string               `json:"columns" jsonschema:"the result column names, in order"`
	Rows      []map[string]any       `json:"rows" jsonschema:"the result rows as objects keyed by column name"`
	RowCount  int                    `json:"row_count" jsonschema:"number of rows returned (after any truncation)"`
	Truncated bool                   `json:"truncated" jsonschema:"true when the result was cut short by a row or size cap"`
	Omitted   []engine.OmittedColumn `json:"omitted,omitempty" jsonschema:"columns dropped because their values are not representable as JSON text, with the reason"`
}

// runSQL executes read-only SQL against a configured database. The query runs in
// a read-only transaction under a statement timeout with row/size caps; binary
// (non-UTF-8) values are omitted and reported in `omitted`.
func (h *Handlers) runSQL(ctx context.Context, _ *mcp.CallToolRequest, in runSQLInput) (*mcp.CallToolResult, runSQLOutput, error) {
	result, err := h.Services.Engine.RunReadOnly(ctx, in.Database, in.SQL)
	if err != nil {
		return nil, runSQLOutput{}, err
	}

	summary := fmt.Sprintf("Returned %d row(s) from %q.", result.RowCount, in.Database)
	if result.Truncated {
		summary += " Result was truncated by a row/size cap."
	}
	if len(result.Omitted) > 0 {
		summary += fmt.Sprintf(" %d column(s) omitted (not representable as JSON text).", len(result.Omitted))
	}

	return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: summary}},
		}, runSQLOutput{
			Columns:   result.Columns,
			Rows:      result.Rows,
			RowCount:  result.RowCount,
			Truncated: result.Truncated,
			Omitted:   result.Omitted,
		}, nil
}

// registerRunSQL adds the run_sql tool. It only reads data (read-only
// transaction), but touches external databases, so OpenWorldHint is true.
func (h *Handlers) registerRunSQL(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "run_sql",
		Description: "Runs read-only SQL against a configured database and returns the rows as JSON. The query executes inside a read-only transaction under a statement timeout, and the result is capped by row count and size (see `truncated`). Binary/non-UTF-8 column values are omitted and listed in `omitted`. Use list_databases to discover database names.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(true),
		},
	}, h.runSQL)
}

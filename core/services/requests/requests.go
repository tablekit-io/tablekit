// Package requests is the audit log for MCP traffic: one mcp_requests row per
// JSON-RPC request the server handles (initialize, tools/list, tools/call,
// resources/read, ...), written by the server's receiving middleware after the
// handler runs. It exists to answer "did this request ever reach the server, and
// what did it return?" — especially for the app-only bridge tools (fetch_chart_data,
// get_export_url) the agent never sees. It stores requests only; it never blocks
// or alters the request it records.
package requests

import (
	"context"
	"database/sql"
	"fmt"

	"core/db/gen/tablekit/public/table"

	"github.com/google/uuid"

	. "github.com/go-jet/jet/v2/postgres"
)

// Entry is one MCP request to record. ToolName is empty for non-tool methods;
// Params/Result/Error are raw JSON (nil = stored as SQL NULL).
type Entry struct {
	Method     string
	ToolName   string
	ClientID   string
	Params     []byte
	Result     []byte
	Error      []byte
	DurationMS int
}

// RequestLog persists MCP request audit rows in the mcp_requests table.
type RequestLog interface {
	Log(ctx context.Context, entry Entry) error
}

type requestLog struct {
	database *sql.DB
}

// New returns a RequestLog over the given database. The schema is owned by the db
// package's migrations; this type only writes rows.
func New(database *sql.DB) RequestLog {
	return &requestLog{database: database}
}

// Log inserts one audit row. The id is a UUIDv7 (time-ordered) so rows sort by
// arrival. JSON columns are sent as text and cast into jsonb by Postgres — the
// same pattern the databases/config stores use.
func (r *requestLog) Log(ctx context.Context, entry Entry) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate request id: %w", err)
	}
	stmt := table.McpRequests.
		INSERT(
			table.McpRequests.ID,
			table.McpRequests.Method,
			table.McpRequests.ToolName,
			table.McpRequests.ClientID,
			table.McpRequests.Params,
			table.McpRequests.Result,
			table.McpRequests.Error,
			table.McpRequests.DurationMs,
		).
		VALUES(
			UUID(id),
			entry.Method,
			nullableText(entry.ToolName),
			nullableUUID(entry.ClientID),
			nullableJSON(entry.Params),
			nullableJSON(entry.Result),
			nullableJSON(entry.Error),
			entry.DurationMS,
		)
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("log mcp request: %w", err)
	}
	return nil
}

// nullableText renders an empty string as SQL NULL, any other value as text.
func nullableText(value string) any {
	if value == "" {
		return NULL
	}
	return value
}

// nullableJSON renders nil bytes as SQL NULL, otherwise the raw JSON as text for
// Postgres to cast into the jsonb column.
func nullableJSON(raw []byte) any {
	if raw == nil {
		return NULL
	}
	return string(raw)
}

// nullableUUID renders an empty or unparseable client id as SQL NULL, otherwise a
// typed uuid value for the client_id column. The audit log is best-effort, so an
// absent/malformed id is stored as NULL rather than failing the write.
func nullableUUID(value string) any {
	id, err := uuid.Parse(value)
	if err != nil {
		return NULL
	}
	return UUID(id)
}

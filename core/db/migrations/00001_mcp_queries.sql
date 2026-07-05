-- +goose Up
-- mcp_queries stores one descriptor per run_query call: the database it targeted,
-- the user's read-only SQL, and the agent's plain-language description. Result rows
-- are NOT stored; retrieve_results / fetch_chart_data / get_export_url re-run the
-- SQL against the live database using this descriptor.
CREATE TABLE mcp_queries (
    id          TEXT PRIMARY KEY,
    database    TEXT NOT NULL,
    sql         TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE mcp_queries;

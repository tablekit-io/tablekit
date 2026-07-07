-- +goose Up
-- Bind stored queries to a physical database_id instead of the databases.yaml
-- name. The old `database` column held the name, which follows a repoint to a
-- different physical database; database_id pins the database the query was
-- actually saved against. Descriptors are transient — the rows are re-run against
-- live data and never store results — so existing rows are discarded rather than
-- backfilled to a physical identity we cannot reconstruct.
DELETE FROM mcp_queries;
ALTER TABLE mcp_queries DROP COLUMN database;
ALTER TABLE mcp_queries ADD COLUMN database_id UUID NOT NULL REFERENCES databases (id);

-- +goose Down
ALTER TABLE mcp_queries DROP COLUMN database_id;
ALTER TABLE mcp_queries ADD COLUMN database TEXT NOT NULL DEFAULT '';

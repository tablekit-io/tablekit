-- +goose Up
-- databases records one row per PHYSICAL database tablekit has ever queried,
-- keyed by a driver-derived fingerprint rather than the databases.yaml name. The
-- name is only a label; its connection details can be repointed at a different
-- database. On the first query against a name the engine connects, derives a
-- stable identity_key (postgres system_identifier + database oid; mysql/mariadb
-- server_uuid + schema), and mints — or matches — a database_id here. mcp_queries
-- binds each stored query to that database_id, so a later repoint is detectable.
CREATE TABLE databases (
    id           UUID PRIMARY KEY,                   -- uuidv7, generated in Go
    name         TEXT NOT NULL,                      -- databases.yaml name last seen under
    type         TEXT NOT NULL,                      -- postgres | mysql | mariadb
    identity_key TEXT NOT NULL UNIQUE,               -- static fingerprint; the match key
    identity     JSONB NOT NULL DEFAULT '{}'::jsonb, -- structured fingerprint, observability only
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Re-run recovers the last-seen name from a database_id, so name lookups are hot.
CREATE INDEX databases_name_idx ON databases (name);

-- +goose Down
DROP TABLE databases;

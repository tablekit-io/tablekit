-- Consolidated tablekit state schema, authored from scratch (pre-launch, no
-- migration history to preserve). Applied verbatim by the 00001_schema.go goose
-- migration. Foreign keys live separately in foreign_keys.sql and are applied
-- only in development — see that file and the migration.

-- client_type distinguishes the two kinds of client that share the clients table:
-- 'oauth' (dynamic OAuth registrations) and 'static' (a CLI-minted static token
-- registers its own client).
CREATE TYPE client_type AS ENUM ('oauth', 'static');

-- clients owns both client kinds. paired_at replaces the old oauth_paired_clients
-- table: a NULL means unpaired, a timestamp records when the client was paired.
CREATE TABLE clients (
    id            UUID PRIMARY KEY,
    client_name   TEXT,
    type          client_type NOT NULL,
    redirect_uris JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    paired_at     TIMESTAMPTZ
);

-- config is the key/value store (pairing_mode lives here). value is jsonb.
CREATE TABLE config (
    key   VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL
);

-- database_type is the engine family of a queried database.
CREATE TYPE database_type AS ENUM ('postgres', 'mysql', 'mariadb');

-- databases records one row per physical database tablekit has queried, keyed by
-- a driver-derived fingerprint rather than the databases.yaml name.
CREATE TABLE databases (
    id           UUID PRIMARY KEY,
    name         TEXT NOT NULL,
    type         database_type NOT NULL,
    identity_key TEXT NOT NULL UNIQUE,
    identity     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX databases_name_idx ON databases (name);

-- queries stores one descriptor per query_database call. Bound to the physical
-- database_id and to the client_id that created it, so every stored query is
-- attributable to a client.
CREATE TABLE queries (
    id          UUID PRIMARY KEY,
    sql         TEXT NOT NULL,
    description TEXT NOT NULL,
    database_id UUID NOT NULL,
    client_id   UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX queries_database_id_idx ON queries (database_id);
CREATE INDEX queries_client_id_idx ON queries (client_id);

-- oauth_auth_codes: the authorization-code grant's short-lived codes. id is the
-- PK; code is the value handed out in the redirect (indexed for redemption).
CREATE TABLE oauth_auth_codes (
    id             UUID PRIMARY KEY,
    code           TEXT NOT NULL,
    client_id      UUID NOT NULL,
    redirect_uri   TEXT NOT NULL,
    code_challenge TEXT NOT NULL,
    scope          TEXT NOT NULL,
    user_id        TEXT NOT NULL,
    expires_at     TIMESTAMPTZ NOT NULL
);
CREATE INDEX oauth_auth_codes_code_idx ON oauth_auth_codes (code);
CREATE INDEX oauth_auth_codes_client_id_idx ON oauth_auth_codes (client_id);

-- oauth_token_chains tracks refresh-token rotation. revoked_at NULL means live.
CREATE TABLE oauth_token_chains (
    id                 UUID PRIMARY KEY,
    client_id          UUID NOT NULL,
    user_id            TEXT NOT NULL,
    scope              TEXT NOT NULL,
    redirect_uri       TEXT NOT NULL,
    revoked_at         TIMESTAMPTZ,
    invalidated_before TIMESTAMPTZ NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX oauth_token_chains_client_id_idx ON oauth_token_chains (client_id);

-- static_tokens are the CLI-minted long-lived tokens (one per static client).
-- revoked_at NULL means live.
CREATE TABLE static_tokens (
    id         UUID PRIMARY KEY,
    client_id  UUID NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX static_tokens_client_id_idx ON static_tokens (client_id);

-- mcp_requests is the audit log of every MCP JSON-RPC request. client_id is
-- nullable: the audit writer is best-effort and must never drop a row.
CREATE TABLE mcp_requests (
    id          UUID PRIMARY KEY,
    method      TEXT NOT NULL,
    tool_name   TEXT,
    client_id   UUID,
    params      JSONB,
    result      JSONB,
    error       JSONB,
    duration_ms INTEGER NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX mcp_requests_tool_name_idx ON mcp_requests (tool_name);
CREATE INDEX mcp_requests_client_id_idx ON mcp_requests (client_id);
CREATE INDEX mcp_requests_created_at_idx ON mcp_requests (created_at);

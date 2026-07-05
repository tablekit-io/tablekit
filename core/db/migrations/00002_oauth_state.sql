-- +goose Up
-- OAuth/MCP server state, moved out of the former clients.json / tokens.json
-- flat files into SQLite. Table names mirror the reference Postgres schema the
-- store package was always modeled on (oauth_clients, oauth_auth_codes,
-- oauth_token_chains). signing.key stays a file — it is a raw secret, not JSON.

-- Registered OAuth clients (RFC 7591) plus CLI bearer clients (type='bearer',
-- null client_name, empty redirect_uris). redirect_uris is a JSON array stored
-- whole: it is only ever read or written as a unit.
CREATE TABLE oauth_clients (
    client_id     TEXT PRIMARY KEY,
    client_name   TEXT,
    type          TEXT NOT NULL DEFAULT '',
    redirect_uris TEXT NOT NULL DEFAULT '[]',
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- One-time, PKCE-bound authorization codes. Deleted on redemption (single use).
CREATE TABLE oauth_auth_codes (
    code           TEXT PRIMARY KEY,
    client_id      TEXT NOT NULL,
    redirect_uri   TEXT NOT NULL,
    code_challenge TEXT NOT NULL,
    scope          TEXT NOT NULL,
    user_id        TEXT NOT NULL,
    expires_at     TIMESTAMP NOT NULL
);

-- Refresh-token lineages. invalidated_before is the rotation cutoff: any refresh
-- token issued at or before it is a replay, which revokes the whole chain.
CREATE TABLE oauth_token_chains (
    id                 TEXT PRIMARY KEY,
    client_id          TEXT NOT NULL,
    user_id            TEXT NOT NULL,
    scope              TEXT NOT NULL,
    redirect_uri       TEXT NOT NULL,
    revoked            BOOLEAN NOT NULL DEFAULT 0,
    invalidated_before TIMESTAMP NOT NULL,
    created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Long-lived CLI-minted bearer tokens, looked up by jti on every MCP request so
-- they can be revoked.
CREATE TABLE oauth_bearer_tokens (
    id         TEXT PRIMARY KEY,
    client_id  TEXT NOT NULL,
    revoked    BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

-- Clients allowed to authenticate. Membership is the paired set.
CREATE TABLE oauth_paired_clients (
    client_id TEXT PRIMARY KEY
);

-- Generic key/value config with JSON-encoded values. Currently holds only
-- 'pairing_mode'; an absent row means the Go default ("once"), matching the old
-- loadClients behavior.
CREATE TABLE config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- +goose Down
DROP TABLE config;
DROP TABLE oauth_paired_clients;
DROP TABLE oauth_bearer_tokens;
DROP TABLE oauth_token_chains;
DROP TABLE oauth_auth_codes;
DROP TABLE oauth_clients;

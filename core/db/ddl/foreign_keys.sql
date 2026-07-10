-- Real foreign-key constraints, applied by 00001_schema.go ONLY when
-- TABLEKIT_ENV=development. In other environments these columns exist and are
-- indexed (see schema.sql) but carry no constraint, so production keeps the
-- flexibility/perf of unconstrained columns while development gets full
-- referential integrity for debugging.
ALTER TABLE queries
    ADD CONSTRAINT queries_database_id_fkey FOREIGN KEY (database_id) REFERENCES databases (id);
ALTER TABLE queries
    ADD CONSTRAINT queries_client_id_fkey FOREIGN KEY (client_id) REFERENCES clients (id);
ALTER TABLE oauth_auth_codes
    ADD CONSTRAINT oauth_auth_codes_client_id_fkey FOREIGN KEY (client_id) REFERENCES clients (id);
ALTER TABLE oauth_token_chains
    ADD CONSTRAINT oauth_token_chains_client_id_fkey FOREIGN KEY (client_id) REFERENCES clients (id);
ALTER TABLE static_tokens
    ADD CONSTRAINT static_tokens_client_id_fkey FOREIGN KEY (client_id) REFERENCES clients (id);
ALTER TABLE mcp_requests
    ADD CONSTRAINT mcp_requests_client_id_fkey FOREIGN KEY (client_id) REFERENCES clients (id);

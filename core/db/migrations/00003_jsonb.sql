-- +goose Up
-- redirect_uris and config.value hold JSON but were declared TEXT — a holdover
-- from the SQLite schema this store was ported from (SQLite has no native JSON
-- type). On Postgres, store them as jsonb so the database validates the JSON and
-- the generated code can type them precisely. The USING casts reinterpret the
-- existing TEXT (already valid JSON) as jsonb in place.
ALTER TABLE oauth_clients
    ALTER COLUMN redirect_uris DROP DEFAULT,
    ALTER COLUMN redirect_uris TYPE jsonb USING redirect_uris::jsonb,
    ALTER COLUMN redirect_uris SET DEFAULT '[]'::jsonb;

ALTER TABLE config
    ALTER COLUMN value TYPE jsonb USING value::jsonb;

-- +goose Down
ALTER TABLE config
    ALTER COLUMN value TYPE text USING value::text;

ALTER TABLE oauth_clients
    ALTER COLUMN redirect_uris DROP DEFAULT,
    ALTER COLUMN redirect_uris TYPE text USING redirect_uris::text,
    ALTER COLUMN redirect_uris SET DEFAULT '[]';

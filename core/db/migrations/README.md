# migrations

The schema is authored from scratch, not as an incremental history. The actual
DDL lives in `../ddl/schema.sql` (tables/indexes, always applied) and
`../ddl/foreign_keys.sql` (foreign keys, applied only when
`TABLEKIT_ENV=development`). Both are executed by the single goose Go migration
in `../00001_schema.go`.

This directory is intentionally otherwise empty: goose requires a migrations
directory to exist, but the migration itself is a registered Go migration rather
than a `.sql` file. This placeholder keeps the embedded directory non-empty.

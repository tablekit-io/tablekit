package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
)

// schemaSQL creates every table, type and index. It is applied on every fresh
// database. Foreign keys are deliberately excluded — see foreignKeysSQL.
//
//go:embed ddl/schema.sql
var schemaSQL string

// foreignKeysSQL adds the real foreign-key constraints. It is applied only when
// TABLE_KIT_ENV=development, so production keeps indexed-but-unconstrained columns
// while development gets full referential integrity.
//
//go:embed ddl/foreign_keys.sql
var foreignKeysSQL string

// developmentEnv is the TABLE_KIT_ENV value that turns on real foreign keys.
const developmentEnv = "development"

// isDevelopment reports whether real foreign keys should be materialized.
func isDevelopment() bool {
	return os.Getenv("TABLE_KIT_ENV") == developmentEnv
}

// init registers the single from-scratch schema migration. It is a Go migration
// (not a .sql file) so the foreign-key step can branch on the environment, which
// SQL migrations cannot do. The filename is passed explicitly so goose derives
// version 1 without relying on runtime.Caller.
func init() {
	goose.AddNamedMigrationContext("00001_schema.go", upSchema, downSchema)
}

// upSchema creates the schema, then materializes foreign keys in development.
func upSchema(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if isDevelopment() {
		if _, err := tx.ExecContext(ctx, foreignKeysSQL); err != nil {
			return fmt.Errorf("apply foreign keys: %w", err)
		}
	}
	return nil
}

// downSchema drops everything the up created. CASCADE clears the foreign keys and
// indexes with their tables; the enum type is dropped last.
func downSchema(ctx context.Context, tx *sql.Tx) error {
	const drop = `
DROP TABLE IF EXISTS mcp_requests, static_tokens, oauth_token_chains,
    oauth_auth_codes, queries, databases, config, clients CASCADE;
DROP TYPE IF EXISTS client_type, database_type;`
	if _, err := tx.ExecContext(ctx, drop); err != nil {
		return fmt.Errorf("drop schema: %w", err)
	}
	return nil
}

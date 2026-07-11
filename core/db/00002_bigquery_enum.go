package db

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

// init registers the migration that adds 'bigquery' to the database_type enum.
// It is a NoTx migration because Postgres forbids ALTER TYPE ... ADD VALUE inside
// a transaction block, and goose wraps ordinary migrations in one. The filename is
// passed explicitly so goose derives version 2 (it runs after 00001_schema.go
// regardless of init order, which goose sorts by version).
func init() {
	goose.AddNamedMigrationNoTxContext("00002_bigquery_enum.go", upBigQueryEnum, downBigQueryEnum)
}

// upBigQueryEnum appends 'bigquery' to the database_type enum. IF NOT EXISTS makes
// it a no-op on a fresh database, where 00001's schema.sql already created the enum
// with the value, and additive on an existing database created before this value
// was introduced. So both converge on the same enum.
func upBigQueryEnum(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `ALTER TYPE database_type ADD VALUE IF NOT EXISTS 'bigquery'`)
	return err
}

// downBigQueryEnum is a no-op: Postgres cannot drop a value from an enum type, so
// there is nothing to reverse. The value is harmless when unused.
func downBigQueryEnum(_ context.Context, _ *sql.DB) error {
	return nil
}

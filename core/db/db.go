// Package db owns the embedded SQLite database that backs tablekit's own state
// (currently the run_query descriptor log). It opens the database under DataDir,
// creating the file if absent, and brings the schema to the latest revision with
// goose on every startup using migrations embedded into the binary — so a fresh
// data directory and an upgraded binary both converge on the current schema with
// no external tooling.
//
// This is tablekit's *own* store, distinct from the user databases the engine
// queries: those are remote Postgres/MySQL/MariaDB; this is a local SQLite file.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // pure-Go driver, registered under the name "sqlite"
)

// migrations holds the goose SQL migrations, embedded so the binary needs no
// migration files on disk. Each is an ordered NNNNN_name.sql with +goose
// Up/Down sections.
//
//go:embed migrations/*.sql
var migrations embed.FS

// dbFileName is the SQLite database file created under DataDir.
const dbFileName = "tablekit.db"

// Open opens (creating if needed) the SQLite database under dataDir and migrates
// it to the latest schema before returning. The returned *sql.DB is ready to use
// and must be closed by the caller on shutdown.
func Open(dataDir string) (*sql.DB, error) {
	// Ensure the data directory exists: this is its first consumer on startup
	// (before the store), and sqlite won't create missing parent directories, so
	// a fresh DATA_DIR would otherwise fail to open the database file.
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("create data dir %s: %w", dataDir, err)
	}
	path := filepath.Join(dataDir, dbFileName)
	// WAL for concurrent readers alongside the single writer; foreign_keys on so
	// constraints are enforced; busy_timeout so a briefly-locked writer waits
	// rather than failing immediately.
	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)"
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}
	if err := migrate(database); err != nil {
		database.Close()
		return nil, err
	}
	return database, nil
}

// migrate applies all pending embedded migrations. It is idempotent: goose tracks
// the applied revisions in its own goose_db_version table, so a second call is a
// no-op.
func migrate(database *sql.DB) error {
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.Up(database, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

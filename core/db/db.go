// Package db owns the Postgres database that backs tablekit's own state (the
// OAuth/MCP server state and the query_database descriptor log). It connects using a
// DSN, waits for the server to accept connections, and brings the schema to the
// latest revision with goose on every startup using migrations embedded into the
// binary — so a fresh database and an upgraded binary both converge on the
// current schema with no external tooling.
//
// This is tablekit's *own* store, distinct from the user databases the engine
// queries: those are arbitrary remote Postgres/MySQL/MariaDB targets configured
// in databases.yaml; this is tablekit's dedicated application database.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver, registered as "pgx"
)

// migrations holds the goose SQL migrations, embedded so the binary needs no
// migration files on disk. Each is an ordered NNNNN_name.sql with +goose
// Up/Down sections.
//
//go:embed migrations/*.sql
var migrations embed.FS

// pingRetries and pingInterval bound the wait for Postgres to start accepting
// connections. On a fresh `docker compose up` the server may not be listening
// the instant core boots even with depends_on health gating, so a short retry
// loop turns a transient dial failure into a clean startup rather than a crash.
const (
	pingRetries  = 40
	pingInterval = 250 * time.Millisecond
)

// Open connects to the Postgres database named by dsn, waits for it to accept
// connections, and migrates it to the latest schema before returning. The
// returned *sql.DB is ready to use and must be closed by the caller on shutdown.
func Open(dsn string) (*sql.DB, error) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := waitReady(database); err != nil {
		database.Close()
		return nil, err
	}
	if err := Migrate(database); err != nil {
		database.Close()
		return nil, err
	}
	return database, nil
}

// waitReady pings the database until it responds or the retry budget is spent,
// so a Postgres that is still starting up doesn't fail startup outright.
func waitReady(database *sql.DB) error {
	var err error
	for attempt := 0; attempt < pingRetries; attempt++ {
		if err = database.Ping(); err == nil {
			return nil
		}
		time.Sleep(pingInterval)
	}
	return fmt.Errorf("postgres not ready after %s: %w", time.Duration(pingRetries)*pingInterval, err)
}

// Migrate applies all pending embedded migrations. It is idempotent: goose tracks
// the applied revisions in its own goose_db_version table, so a second call is a
// no-op. It is exported so tests can migrate an isolated database directly.
func Migrate(database *sql.DB) error {
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.Up(database, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

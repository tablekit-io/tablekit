package store

import (
	"database/sql"
	"os"
	"testing"

	"core/db/dbtest"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available), so each test gets an isolated migrated database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

// newDB returns a fresh migrated Postgres database for a test. Each repository is
// built over it directly; a second repository over the same handle stands in for
// a process restart (state persists in Postgres, not the repository).
func newDB(t *testing.T) *sql.DB {
	t.Helper()
	return dbtest.New(t)
}

package store

import (
	"os"
	"testing"

	"core/db/dbtest"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available), so each test gets an isolated migrated database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

// newStore returns a Store backed by a fresh migrated Postgres database (for the
// oauth_* tables).
func newStore(t *testing.T) *Store {
	t.Helper()
	return New(dbtest.New(t))
}

// reopen returns a fresh Store over the same database, to assert state persists
// across process restarts.
func reopen(t *testing.T, s *Store) *Store {
	t.Helper()
	return New(s.database)
}

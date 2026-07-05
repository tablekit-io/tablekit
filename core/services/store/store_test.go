package store

import (
	"os"
	"testing"

	"core/db/dbtest"

	"github.com/stretchr/testify/require"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available), so each test gets an isolated migrated database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

// newStore returns a Store backed by a fresh migrated Postgres database (for the
// oauth_* tables) and a temp dir (for signing.key).
func newStore(t *testing.T) *Store {
	t.Helper()
	database := dbtest.New(t)
	storageService, err := New(t.TempDir(), database)
	require.NoError(t, err)
	return storageService
}

// reopen returns a fresh Store over the same database and directory, to assert
// state persists across process restarts.
func reopen(t *testing.T, s *Store) *Store {
	t.Helper()
	reopened, err := New(s.directory, s.database)
	require.NoError(t, err)
	return reopened
}

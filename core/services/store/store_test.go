package store

import (
	"testing"

	"core/db"

	"github.com/stretchr/testify/require"
)

// newStore returns a Store backed by a fresh migrated SQLite database (for the
// oauth_* tables) and a temp dir (for signing.key). The database is closed when
// the test ends.
func newStore(t *testing.T) *Store {
	t.Helper()
	database, err := db.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
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

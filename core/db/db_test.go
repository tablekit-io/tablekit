package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenCreatesMissingDataDir: a fresh, not-yet-existing data directory is
// created on Open rather than failing — sqlite won't create missing parents, and
// db.Open is the first startup consumer of DATA_DIR.
func TestOpenCreatesMissingDataDir(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "nested", "data")

	database, err := Open(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })

	assert.DirExists(t, dataDir)
	assert.FileExists(t, filepath.Join(dataDir, dbFileName))

	// Migrations ran: the descriptor table is queryable.
	require.NoError(t, database.Ping())
	_, err = database.Exec(`SELECT 1`)
	require.NoError(t, err)
}

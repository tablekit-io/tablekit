package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newStore returns a Store backed by a fresh temp dir.
func newStore(t *testing.T) *Store {
	t.Helper()
	storageService, err := New(t.TempDir())
	require.NoError(t, err)
	return storageService
}

func TestNewFailsOnCorruptState(t *testing.T) {
	directory := t.TempDir()
	// Client object missing the required client_id field.
	bad := `{"clients":{"x":{"redirect_uris":["http://x/cb"],"created_at":"2026-01-01T00:00:00Z"}}}`
	require.NoError(t, os.WriteFile(filepath.Join(directory, "clients.json"), []byte(bad), 0o600))

	_, err := New(directory)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")
}

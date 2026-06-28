package queries_test

import (
	"context"
	"testing"

	"core/db"
	"core/services/queries"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRepository opens a migrated SQLite database in a temp dir and returns a
// repository over it. The database is closed when the test ends.
func newRepository(t *testing.T) *queries.Repository {
	t.Helper()
	database, err := db.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	return queries.New(database)
}

func TestSaveThenGet(t *testing.T) {
	repository := newRepository(t)
	ctx := context.Background()

	key, err := repository.Save(ctx, "emerald", "SELECT 1", "a trivial probe")
	require.NoError(t, err)
	require.NotEmpty(t, key)

	got, err := repository.Get(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, key, got.ID)
	assert.Equal(t, "emerald", got.Database)
	assert.Equal(t, "SELECT 1", got.SQL)
	assert.Equal(t, "a trivial probe", got.Description)
	assert.False(t, got.CreatedAt.IsZero(), "created_at should be set by the default")
}

func TestGetUnknownKeyReturnsNilNil(t *testing.T) {
	repository := newRepository(t)

	got, err := repository.Get(context.Background(), "does-not-exist")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSaveGeneratesDistinctKeys(t *testing.T) {
	repository := newRepository(t)
	ctx := context.Background()

	first, err := repository.Save(ctx, "emerald", "SELECT 1", "one")
	require.NoError(t, err)
	second, err := repository.Save(ctx, "emerald", "SELECT 1", "two")
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

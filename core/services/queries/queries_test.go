package queries_test

import (
	"context"
	"os"
	"testing"

	"core/db/dbtest"
	"core/services/queries"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available), so each test gets an isolated migrated database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

// newRepository opens a fresh migrated Postgres database and returns a repository
// over it. The database is dropped when the test ends.
func newRepository(t *testing.T) queries.QueryRepository {
	t.Helper()
	return queries.New(dbtest.New(t))
}

func TestSaveThenGet(t *testing.T) {
	repository := newRepository(t)
	ctx := context.Background()

	key, err := repository.Save(ctx, "cafe", "SELECT 1", "a trivial probe")
	require.NoError(t, err)
	require.NotEmpty(t, key)

	got, err := repository.Get(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, key, got.ID)
	assert.Equal(t, "cafe", got.Database)
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

	first, err := repository.Save(ctx, "cafe", "SELECT 1", "one")
	require.NoError(t, err)
	second, err := repository.Save(ctx, "cafe", "SELECT 1", "two")
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

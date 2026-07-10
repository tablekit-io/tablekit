package queries_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"core/db/dbtest"
	"core/services/databases"
	"core/services/queries"
	"core/services/store"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available), so each test gets an isolated migrated database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

// newRepository opens a fresh migrated Postgres database and returns a queries
// repository plus a database_id and client_id that satisfy the queries FKs. The
// database is dropped when the test ends.
func newRepository(t *testing.T) (repo queries.QueryRepository, databaseID, clientID uuid.UUID) {
	t.Helper()
	database := dbtest.New(t)
	return queries.New(database), seedDatabase(t, database, "cafe"), seedClient(t, database)
}

// seedClient inserts a client row (queries.client_id references it) and returns
// its id.
func seedClient(t *testing.T, database *sql.DB) uuid.UUID {
	t.Helper()
	id, err := uuid.NewV7()
	require.NoError(t, err)
	require.NoError(t, store.NewClientRepository(database).SaveClient(context.Background(), &store.Client{
		ClientID:     id,
		Type:         store.ClientTypeStatic,
		RedirectURIs: []string{},
		CreatedAt:    time.Now(),
	}))
	return id
}

// seedDatabase inserts a databases row (queries.database_id references it) and
// returns its id.
func seedDatabase(t *testing.T, database *sql.DB, name string) uuid.UUID {
	t.Helper()
	newID, err := uuid.NewV7()
	require.NoError(t, err)
	id, err := databases.NewRepository(database).Upsert(context.Background(), newID, databases.Record{
		Name:        name,
		Type:        "postgres",
		IdentityKey: "pg-test-" + name,
		Identity:    map[string]string{"database_name": name},
	})
	require.NoError(t, err)
	return id
}

func TestSaveThenGet(t *testing.T) {
	repository, databaseID, clientID := newRepository(t)
	ctx := context.Background()

	key, err := repository.Save(ctx, databaseID, clientID, "SELECT 1", "a trivial probe")
	require.NoError(t, err)
	require.NotEmpty(t, key)

	got, err := repository.Get(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, key, got.ID)
	assert.Equal(t, databaseID, got.DatabaseID)
	assert.Equal(t, "SELECT 1", got.SQL)
	assert.Equal(t, "a trivial probe", got.Description)
	assert.False(t, got.CreatedAt.IsZero(), "created_at should be set by the default")
}

func TestGetUnknownKeyReturnsNilNil(t *testing.T) {
	repository, _, _ := newRepository(t)

	got, err := repository.Get(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSaveGeneratesDistinctKeys(t *testing.T) {
	repository, databaseID, clientID := newRepository(t)
	ctx := context.Background()

	first, err := repository.Save(ctx, databaseID, clientID, "SELECT 1", "one")
	require.NoError(t, err)
	second, err := repository.Save(ctx, databaseID, clientID, "SELECT 1", "two")
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

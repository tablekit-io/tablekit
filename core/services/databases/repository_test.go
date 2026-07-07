package databases_test

import (
	"context"
	"os"
	"testing"

	"core/db/dbtest"
	"core/services/databases"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available), so each test gets an isolated migrated database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

func newRepository(t *testing.T) databases.Repository {
	t.Helper()
	return databases.NewRepository(dbtest.New(t))
}

func newID(t *testing.T) uuid.UUID {
	t.Helper()
	id, err := uuid.NewV7()
	require.NoError(t, err)
	return id
}

func TestUpsertInsertsThenGet(t *testing.T) {
	repository := newRepository(t)
	ctx := context.Background()

	id := newID(t)
	stored, err := repository.Upsert(ctx, id, databases.Record{
		Name:        "cafe",
		Type:        "postgres",
		IdentityKey: "pg-cluster-1-16384",
		Identity:    map[string]string{"system_identifier": "cluster-1", "database_oid": "16384"},
	})
	require.NoError(t, err)
	assert.Equal(t, id, stored)

	got, err := repository.Get(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "cafe", got.Name)
	assert.Equal(t, "postgres", got.Type)
	assert.Equal(t, "pg-cluster-1-16384", got.IdentityKey)
	assert.Equal(t, "16384", got.Identity["database_oid"])
	assert.False(t, got.CreatedAt.IsZero())
}

func TestUpsertSameIdentityKeyKeepsStableIDAndUpdatesName(t *testing.T) {
	repository := newRepository(t)
	ctx := context.Background()

	firstID, err := repository.Upsert(ctx, newID(t), databases.Record{
		Name: "cafe", Type: "postgres", IdentityKey: "pg-cluster-1", Identity: map[string]string{"v": "1"},
	})
	require.NoError(t, err)

	// A second upsert for the SAME physical database (same identity_key) under a
	// different name and a different candidate id must return the original id.
	secondID, err := repository.Upsert(ctx, newID(t), databases.Record{
		Name: "cafe_renamed", Type: "postgres", IdentityKey: "pg-cluster-1", Identity: map[string]string{"v": "2"},
	})
	require.NoError(t, err)
	assert.Equal(t, firstID, secondID, "one physical database keeps one stable id")

	got, err := repository.Get(ctx, firstID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "cafe_renamed", got.Name, "the last-seen name is stored")
	assert.Equal(t, "2", got.Identity["v"], "the identity is refreshed")
}

func TestGetUnknownIDReturnsNilNil(t *testing.T) {
	repository := newRepository(t)

	got, err := repository.Get(context.Background(), newID(t))
	require.NoError(t, err)
	assert.Nil(t, got)
}

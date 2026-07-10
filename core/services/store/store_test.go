package store

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"core/db/dbtest"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
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

// seedClient inserts a static client and returns its id. Rows that reference a
// client (auth codes, chains, static tokens) need one to exist so the
// development foreign keys are satisfied.
func seedClient(t *testing.T, database *sql.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, NewClientRepository(database).SaveClient(context.Background(), &Client{
		ClientID:     id,
		ClientName:   nil,
		RedirectURIs: []string{},
		Type:         ClientTypeStatic,
		CreatedAt:    time.Now(),
	}))
	return id
}

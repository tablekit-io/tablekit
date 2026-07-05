package db_test

import (
	"os"
	"testing"

	"core/db"
	"core/db/dbtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available).
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

// TestMigrateBringsSchemaUp: a fresh database has the migrations applied, so the
// state tables exist and are queryable. dbtest.New already runs db.Migrate; this
// asserts the schema it produces is what the store code expects.
func TestMigrateBringsSchemaUp(t *testing.T) {
	database := dbtest.New(t)

	// Migrate again to prove idempotence (goose tracks applied revisions).
	require.NoError(t, db.Migrate(database))

	for _, table := range []string{
		"mcp_queries", "oauth_clients", "oauth_auth_codes", "oauth_token_chains",
		"oauth_bearer_tokens", "oauth_paired_clients", "config",
	} {
		var count int
		err := database.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count)
		require.NoErrorf(t, err, "table %s should exist and be queryable", table)
		assert.Zero(t, count, "fresh %s should be empty", table)
	}
}

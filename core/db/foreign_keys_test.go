package db_test

import (
	"database/sql"
	"testing"

	"core/db/dbtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// foreignKeyCount returns how many real foreign-key constraints exist in the
// database's public schema.
func foreignKeyCount(t *testing.T, database *sql.DB) int {
	t.Helper()
	var n int
	require.NoError(t, database.QueryRow(
		`SELECT count(*) FROM pg_constraint WHERE contype = 'f'`).Scan(&n))
	return n
}

// TestForeignKeysMaterializedInDevelopment: when TABLEKIT_ENV=development the
// schema migration also applies foreign_keys.sql, so every FK column carries a
// real constraint. dbtest.New migrates a fresh database reading the env at
// migration time, so setting it here controls the outcome.
func TestForeignKeysMaterializedInDevelopment(t *testing.T) {
	t.Setenv("TABLEKIT_ENV", "development")
	database := dbtest.New(t)
	// queries(database_id, client_id), oauth_auth_codes, oauth_token_chains,
	// static_tokens, mcp_requests -> clients/databases: six constraints.
	assert.Equal(t, 6, foreignKeyCount(t, database))
}

// TestForeignKeysSkippedOutsideDevelopment: with any other TABLEKIT_ENV the FK
// columns exist and are indexed, but no constraint is created.
func TestForeignKeysSkippedOutsideDevelopment(t *testing.T) {
	t.Setenv("TABLEKIT_ENV", "")
	database := dbtest.New(t)
	assert.Zero(t, foreignKeyCount(t, database))
}

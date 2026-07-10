package requests_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"core/db/dbtest"
	"core/services/requests"
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

// readOnly reads the single audit row back as raw text, so the test does not
// depend on the generated model's nullable-jsonb field types.
type readOnly struct {
	method   string
	toolName sql.NullString
	clientID sql.NullString
	params   sql.NullString
	result   sql.NullString
	errCol   sql.NullString
	duration int
}

func readSingle(t *testing.T, database *sql.DB) readOnly {
	t.Helper()
	var row readOnly
	err := database.QueryRowContext(context.Background(),
		`SELECT method, tool_name, client_id, params::text, result::text, error::text, duration_ms FROM mcp_requests`).
		Scan(&row.method, &row.toolName, &row.clientID, &row.params, &row.result, &row.errCol, &row.duration)
	require.NoError(t, err)
	return row
}

// seedClient inserts a client row so a non-null mcp_requests.client_id satisfies
// the development foreign key, and returns its id.
func seedClient(t *testing.T, database *sql.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, store.NewClientRepository(database).SaveClient(context.Background(), &store.Client{
		ClientID:     id,
		Type:         store.ClientTypeStatic,
		RedirectURIs: []string{},
		CreatedAt:    time.Now(),
	}))
	return id
}

func TestLogToolCall(t *testing.T) {
	database := dbtest.New(t)
	clientID := seedClient(t, database)
	log := requests.New(database)

	err := log.Log(context.Background(), requests.Entry{
		Method:     "tools/call",
		ToolName:   "fetch_chart_data",
		ClientID:   clientID.String(),
		Params:     []byte(`{"query_key": "abc"}`),
		Result:     []byte(`{"row_count": 3}`),
		DurationMS: 12,
	})
	require.NoError(t, err)

	row := readSingle(t, database)
	assert.Equal(t, "tools/call", row.method)
	assert.Equal(t, "fetch_chart_data", row.toolName.String)
	assert.Equal(t, clientID.String(), row.clientID.String)
	assert.JSONEq(t, `{"query_key": "abc"}`, row.params.String)
	assert.JSONEq(t, `{"row_count": 3}`, row.result.String)
	assert.False(t, row.errCol.Valid, "error should be NULL on success")
	assert.Equal(t, 12, row.duration)
}

func TestLogNonToolMethodWithError(t *testing.T) {
	database := dbtest.New(t)
	log := requests.New(database)

	// A non-tool method that failed: no tool_name, no result, structured error.
	err := log.Log(context.Background(), requests.Entry{
		Method:     "initialize",
		Error:      []byte(`{"message": "boom"}`),
		DurationMS: 1,
	})
	require.NoError(t, err)

	row := readSingle(t, database)
	assert.Equal(t, "initialize", row.method)
	assert.False(t, row.toolName.Valid, "tool_name should be NULL for non-tool methods")
	assert.False(t, row.clientID.Valid, "client_id should be NULL when empty")
	assert.False(t, row.params.Valid, "params should be NULL when nil")
	assert.False(t, row.result.Valid, "result should be NULL on error")
	assert.JSONEq(t, `{"message": "boom"}`, row.errCol.String)
}

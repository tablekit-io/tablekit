package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"core/e2e/harness"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// queryDatabaseResult mirrors the query_database tool's structured output for
// decoding. Rows are present because callQueryDatabase sets include_results.
type queryDatabaseResult struct {
	ResultKey string `json:"result_key"`
	Columns   []struct {
		Name string `json:"name"`
	} `json:"columns"`
	Rows     []map[string]any `json:"rows"`
	RowCount int              `json:"row_count"`
	HasMore  bool             `json:"has_more"`
}

// columnNames extracts the column names from a query_database result.
func (r queryDatabaseResult) columnNames() []string {
	names := make([]string, len(r.Columns))
	for i, c := range r.Columns {
		names[i] = c.Name
	}
	return names
}

// callQueryDatabase invokes query_database (with the first page of rows inlined)
// and returns the decoded result, plus whether the call was an error (transport
// error or tool IsError).
func callQueryDatabase(t *testing.T, session *mcp.ClientSession, database, query string) (queryDatabaseResult, bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "query_database",
		Arguments: map[string]any{
			"database":        database,
			"sql":             query,
			"description":     "e2e test query",
			"include_results": true,
		},
	})
	if err != nil {
		return queryDatabaseResult{}, true
	}
	if result.IsError {
		return queryDatabaseResult{}, true
	}
	var decoded queryDatabaseResult
	raw, err := json.Marshal(result.StructuredContent)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &decoded))
	return decoded, false
}

// dbCase parameterizes the engine under test.
type dbCase struct {
	engine        string // databases.yaml type
	start         func(t *testing.T) string
	database      string // target database name in details
	username      string
	password      string
	port          int
	seededTable   string // a real seeded table to count
	truncateQuery string // a query returning > the row cap
}

func dbCases() []dbCase {
	return []dbCase{
		{
			engine: "postgres", start: startPostgres,
			database: "cafe", username: "postgres", password: "pw", port: 5432,
			seededTable:   "customers",
			truncateQuery: "SELECT generate_series(1, 3000) AS n",
		},
		{
			engine: "mysql", start: startMySQL,
			database: "dbctx_test_dira", username: "root", password: "pw", port: 3306,
			seededTable:   "users",
			truncateQuery: "SELECT 1 AS n FROM information_schema.columns a CROSS JOIN information_schema.columns b LIMIT 3000",
		},
	}
}

// writeDatabasesYAML writes a one-database config (optionally tunneled) and
// returns its path.
func writeDatabasesYAML(t *testing.T, c dbCase, dbHost, sshBlock string) string {
	t.Helper()
	yaml := fmt.Sprintf(`databases:
  target:
    type: %s
    details:
      host: %s
      port: %d
      database: %s
      username: %s
      password: pw
    tls:
      mode: disable
%s`, c.engine, dbHost, c.port, c.database, c.username, sshBlock)
	path := filepath.Join(t.TempDir(), "databases.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
	return path
}

// runMatrix exercises query_database/list_available_databases against a started server.
func runMatrix(t *testing.T, c dbCase, configPath string) {
	t.Helper()
	server := harness.StartServerEnv(t, "DATABASES_FILE="+configPath)
	_, token := harness.GenerateToken(t, server)
	session, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	// list_available_databases returns the configured target.
	ctx := context.Background()
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_available_databases"})
	require.NoError(t, err)
	require.False(t, listResult.IsError)
	listed := listResult.StructuredContent.(map[string]any)
	databases := listed["databases"].([]any)
	require.Len(t, databases, 1)
	first := databases[0].(map[string]any)
	assert.Equal(t, "target", first["name"])
	assert.Equal(t, c.engine, first["type"])

	// Seeded SELECT correctness: a real table is queryable.
	count, isErr := callQueryDatabase(t, session, "target", "SELECT count(*) AS n FROM "+c.seededTable)
	require.False(t, isErr, "count query on %s should succeed", c.seededTable)
	require.Equal(t, 1, count.RowCount)
	assert.Contains(t, count.columnNames(), "n")

	// Typed literal round-trip: columns + values arrive intact.
	lit, isErr := callQueryDatabase(t, session, "target", "SELECT 7 AS answer, 'tablekit' AS name")
	require.False(t, isErr)
	require.Len(t, lit.Rows, 1)
	assert.Equal(t, "tablekit", lit.Rows[0]["name"])

	// Read-only rejection: a DML write fails inside the read-only transaction.
	// (INSERT ... SELECT * keeps it column-agnostic; the read-only error fires
	// before row evaluation. DDL is not used here: MySQL DDL implicitly commits,
	// so it sidesteps the transaction — only DML is reliably blocked.)
	writeQuery := "INSERT INTO " + c.seededTable + " SELECT * FROM " + c.seededTable
	_, isErr = callQueryDatabase(t, session, "target", writeQuery)
	assert.True(t, isErr, "DML write must be rejected by the read-only transaction")

	// Paging: a result larger than the first page is capped to default_limit
	// (128) with has_more set, rather than returned whole.
	trunc, isErr := callQueryDatabase(t, session, "target", c.truncateQuery)
	require.False(t, isErr)
	assert.True(t, trunc.HasMore)
	assert.Equal(t, 128, trunc.RowCount)

	// Unknown database name returns a clean error.
	_, isErr = callQueryDatabase(t, session, "does-not-exist", "SELECT 1")
	assert.True(t, isErr, "unknown database must error")
}

// TestDatabasesDirect: query_database against postgres and mysql over a direct connection.
func TestDatabasesDirect(t *testing.T) {
	harness.RequireDocker(t)
	for _, c := range dbCases() {
		t.Run(c.engine, func(t *testing.T) {
			t.Parallel()
			dbHost := c.start(t)
			configPath := writeDatabasesYAML(t, c, dbHost, "")
			runMatrix(t, c, configPath)
		})
	}
}

// TestDatabasesOverSSH: query_database against postgres and mysql through the SSH bastion.
func TestDatabasesOverSSH(t *testing.T) {
	harness.RequireDocker(t)
	for _, c := range dbCases() {
		t.Run(c.engine, func(t *testing.T) {
			t.Parallel()
			dbHost := c.start(t)

			authorizedKey, privatePEM := generateSSHKey(t)
			bastion := startBastion(t, authorizedKey)
			keyPath := filepath.Join(t.TempDir(), "id_ed25519")
			require.NoError(t, os.WriteFile(keyPath, privatePEM, 0o600))

			sshBlock := fmt.Sprintf(`    ssh:
      host: %s
      port: 22
      username: root
      sshKeyFilePath: %s
`, bastion, keyPath)
			configPath := writeDatabasesYAML(t, c, dbHost, sshBlock)
			runMatrix(t, c, configPath)
		})
	}
}

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

// runSQLResult mirrors the run_sql tool's structured output for decoding.
type runSQLResult struct {
	Columns   []string         `json:"columns"`
	Rows      []map[string]any `json:"rows"`
	RowCount  int              `json:"row_count"`
	Truncated bool             `json:"truncated"`
}

// callRunSQL invokes run_sql and returns the decoded result, plus whether the
// call was an error (transport error or tool IsError).
func callRunSQL(t *testing.T, session *mcp.ClientSession, database, query string) (runSQLResult, bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "run_sql",
		Arguments: map[string]any{"database": database, "sql": query},
	})
	if err != nil {
		return runSQLResult{}, true
	}
	if result.IsError {
		return runSQLResult{}, true
	}
	var decoded runSQLResult
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
			database: "emerald", username: "postgres", password: "pw", port: 5432,
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

// runMatrix exercises run_sql/list_databases against a started server.
func runMatrix(t *testing.T, c dbCase, configPath string) {
	t.Helper()
	server := harness.StartServerEnv(t, "DATABASES_FILE="+configPath)
	_, token := harness.GenerateToken(t, server)
	session, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	// list_databases returns the configured target.
	ctx := context.Background()
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_databases"})
	require.NoError(t, err)
	require.False(t, listResult.IsError)
	listed := listResult.StructuredContent.(map[string]any)
	databases := listed["databases"].([]any)
	require.Len(t, databases, 1)
	first := databases[0].(map[string]any)
	assert.Equal(t, "target", first["name"])
	assert.Equal(t, c.engine, first["type"])

	// Seeded SELECT correctness: a real table is queryable.
	count, isErr := callRunSQL(t, session, "target", "SELECT count(*) AS n FROM "+c.seededTable)
	require.False(t, isErr, "count query on %s should succeed", c.seededTable)
	require.Equal(t, 1, count.RowCount)
	assert.Contains(t, count.Columns, "n")

	// Typed literal round-trip: columns + values arrive intact.
	lit, isErr := callRunSQL(t, session, "target", "SELECT 7 AS answer, 'tablekit' AS name")
	require.False(t, isErr)
	require.Len(t, lit.Rows, 1)
	assert.Equal(t, "tablekit", lit.Rows[0]["name"])

	// Read-only rejection: a DML write fails inside the read-only transaction.
	// (INSERT ... SELECT * keeps it column-agnostic; the read-only error fires
	// before row evaluation. DDL is not used here: MySQL DDL implicitly commits,
	// so it sidesteps the transaction — only DML is reliably blocked.)
	writeQuery := "INSERT INTO " + c.seededTable + " SELECT * FROM " + c.seededTable
	_, isErr = callRunSQL(t, session, "target", writeQuery)
	assert.True(t, isErr, "DML write must be rejected by the read-only transaction")

	// Truncation: a result larger than the row cap is cut to 2048.
	trunc, isErr := callRunSQL(t, session, "target", c.truncateQuery)
	require.False(t, isErr)
	assert.True(t, trunc.Truncated)
	assert.Equal(t, 2048, trunc.RowCount)

	// Unknown database name returns a clean error.
	_, isErr = callRunSQL(t, session, "does-not-exist", "SELECT 1")
	assert.True(t, isErr, "unknown database must error")
}

// TestDatabasesDirect: run_sql against postgres and mysql over a direct connection.
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

// TestDatabasesOverSSH: run_sql against postgres and mysql through the SSH bastion.
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

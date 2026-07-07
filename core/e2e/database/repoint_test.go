package database

import (
	"context"
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

// writeTargetYAML writes a one-database config named "target" pointing at host.
func writeTargetYAML(t *testing.T, path, host string) {
	t.Helper()
	yaml := fmt.Sprintf(`databases:
  target:
    type: postgres
    details:
      host: %s
      port: 5432
      database: cafe
      username: postgres
      password: pw
    tls:
      mode: disable
`, host)
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
}

// callReadResults invokes read_results for key and reports whether it errored
// (transport error or tool IsError).
func callReadResults(t *testing.T, session *mcp.ClientSession, key string) bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "read_results",
		Arguments: map[string]any{"key": key},
	})
	if err != nil {
		return true
	}
	return result.IsError
}

// TestRepointedDatabaseRefusesStoredQuery is the end-to-end guard: a query saved
// against one physical database must be refused after the same name is repointed
// at a different physical database, while a fresh query against the new database
// still succeeds.
func TestRepointedDatabaseRefusesStoredQuery(t *testing.T) {
	harness.RequireDocker(t)

	hostA := startPostgres(t)
	hostB := startPostgres(t)

	configPath := filepath.Join(t.TempDir(), "databases.yaml")
	writeTargetYAML(t, configPath, hostA)

	server := harness.StartServerEnv(t, "DATABASES_FILE="+configPath)
	_, token := harness.GenerateToken(t, server)
	session, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	// Save a query against physical database A.
	saved, isErr := callQueryDatabase(t, session, "target", "SELECT count(*) AS n FROM customers")
	require.False(t, isErr)
	key := saved.ResultKey
	require.NotEmpty(t, key)

	// read_results works while the name still points at A.
	assert.False(t, callReadResults(t, session, key), "read_results should work before the repoint")

	// Repoint the name at a DIFFERENT physical database (B). The file watcher
	// reloads and invalidates the identity cache, so the next re-run re-derives.
	writeTargetYAML(t, configPath, hostB)

	// The reload is asynchronous; poll until read_results refuses the stale query.
	require.Eventually(t, func() bool {
		return callReadResults(t, session, key)
	}, 15*time.Second, 250*time.Millisecond, "read_results must refuse once the name is repointed")

	// A fresh query against the repointed name mints a new database_id for B and
	// succeeds.
	fresh, isErr := callQueryDatabase(t, session, "target", "SELECT count(*) AS n FROM customers")
	require.False(t, isErr, "a fresh query against the repointed name should succeed")
	assert.NotEmpty(t, fresh.ResultKey)
}

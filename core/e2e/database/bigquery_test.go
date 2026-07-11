package database

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"core/e2e/harness"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bigQueryEmulatorEnv opts a run into the BigQuery emulator e2e. It is off by
// default: the emulator image is heavy to pull and build, and the suite otherwise
// covers BigQuery through unit tests, so this only runs when explicitly requested.
const bigQueryEmulatorEnv = "TABLEKIT_E2E_BIGQUERY"

// requireBigQueryEmulator skips unless docker is available AND the opt-in env is
// set, so the emulator only starts when a run asks for it.
func requireBigQueryEmulator(t *testing.T) {
	t.Helper()
	harness.RequireDocker(t)
	if os.Getenv(bigQueryEmulatorEnv) == "" {
		t.Skipf("set %s=1 to run the BigQuery emulator e2e", bigQueryEmulatorEnv)
	}
}

// startBigQueryEmulator builds (once) and starts the goccy BigQuery emulator with
// the baked seed, and returns its container name and the HTTP endpoint the server
// reaches it on over the shared network.
func startBigQueryEmulator(t *testing.T) (name, endpoint string) {
	t.Helper()
	harness.EnsureImage(t, "tablekit-e2e-bigquery:latest", filepath.Join(dbDir(), "containers", "bigquery"))
	name = harness.RunContainer(t, harness.ContainerSpec{
		Name:  harness.UniqueName("bq"),
		Image: "tablekit-e2e-bigquery:latest",
		Cmd:   []string{"--project=tablekit-test", "--data-from-yaml=/seed.yaml"},
	})
	waitForContainerLog(t, name, 60*time.Second, "9050")
	return name, "http://" + name + ":9050"
}

// waitForContainerLog polls a container's logs until they contain marker, then
// returns. The emulator image is distroless, so an in-container exec probe is not
// possible; its logs are read from the host instead. On timeout it dumps the logs.
func waitForContainerLog(t *testing.T, name string, timeout time.Duration, marker string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		logs, _ := exec.Command("docker", "logs", name).CombinedOutput()
		if strings.Contains(string(logs), marker) {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	logs, _ := exec.Command("docker", "logs", name).CombinedOutput()
	t.Fatalf("emulator %s did not log %q within %s; logs:\n%s", name, marker, timeout, logs)
}

// writeBigQueryYAML writes a one-database bigquery config and returns its path. The
// credentials file is a dummy: the driver talks to the emulator (via
// BIGQUERY_EMULATOR_HOST) with authentication disabled, so the file is never read,
// but the schema requires the path to be present.
func writeBigQueryYAML(t *testing.T) string {
	t.Helper()
	dummyKey := filepath.Join(t.TempDir(), "key.json")
	require.NoError(t, os.WriteFile(dummyKey, []byte("{}"), 0o600))
	yaml := "databases:\n" +
		"  target:\n" +
		"    type: bigquery\n" +
		"    details:\n" +
		"      projectId: tablekit-test\n" +
		"      credentialsFilePath: " + dummyKey + "\n"
	path := filepath.Join(t.TempDir(), "databases.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
	return path
}

// TestBigQueryQueryDatabase drives query_database/list_available_databases against
// the BigQuery emulator: the read-only SELECT path, pagination, the cost hint, and
// the dry-run read-only refusal. Opt-in via TABLEKIT_E2E_BIGQUERY.
func TestBigQueryQueryDatabase(t *testing.T) {
	requireBigQueryEmulator(t)
	_, endpoint := startBigQueryEmulator(t)
	configPath := writeBigQueryYAML(t)

	server := harness.StartServerEnv(t, "DATABASES_FILE="+configPath, "BIGQUERY_EMULATOR_HOST="+endpoint)
	_, token := harness.GenerateToken(t, server)
	session, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	ctx := context.Background()

	// list_available_databases returns the bigquery target and the cost hint.
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_available_databases"})
	require.NoError(t, err)
	require.False(t, listResult.IsError)
	listed := listResult.StructuredContent.(map[string]any)
	databases := listed["databases"].([]any)
	require.Len(t, databases, 1)
	assert.Equal(t, "bigquery", databases[0].(map[string]any)["type"])
	hints, _ := listed["hints_for_ai_agents"].([]any)
	require.NotEmpty(t, hints, "a bigquery database must surface the cost hint")

	// Seeded SELECT: the three seed rows are counted.
	count, isErr := callQueryDatabase(t, session, "target", "SELECT count(*) AS n FROM sample.users")
	require.False(t, isErr, "count over the seeded table should succeed")
	require.Len(t, count.Rows, 1)
	assert.EqualValues(t, 3, count.Rows[0]["n"])

	// Paging: a result larger than the first page (128) is capped with has_more.
	trunc, isErr := callQueryDatabase(t, session, "target", "SELECT n FROM UNNEST(GENERATE_ARRAY(1, 3000)) AS n")
	require.False(t, isErr)
	assert.True(t, trunc.HasMore)
	assert.Equal(t, 128, trunc.RowCount)

	// Read-only: a write is refused by the dry-run statement-type gate.
	_, isErr = callQueryDatabase(t, session, "target",
		"INSERT INTO sample.users (id, name) VALUES (4, 'mallory')")
	assert.True(t, isErr, "a write must be rejected by the read-only gate")
}

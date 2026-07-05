// Package dbtest gives tests a real, migrated Postgres to run against, matching
// production (tablekit's own state lives in Postgres). It starts ONE throwaway
// postgres:17 container per test binary and hands each test its own freshly
// migrated database on it, dropped when the test ends.
//
// Like the engine e2e suite, the container is reached by its DNS name on the
// shared docker network (harness.DockerNetwork), so the test process must sit on
// that network — i.e. run inside the dev container. Where it does not (a bare
// host), New skips the test via harness.RequireDocker.
//
// Usage: a test package calls dbtest.Main from TestMain (so the container spans
// the whole binary), then dbtest.New(t) per test:
//
//	func TestMain(m *testing.M) { os.Exit(dbtest.Main(m)) }
//	func TestX(t *testing.T)     { database := dbtest.New(t); ... }
package dbtest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os/exec"
	"sync"
	"testing"
	"time"

	"core/db"
	"core/e2e/harness"

	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx"
)

// adminPassword is the postgres superuser password the throwaway container runs
// with; it never leaves the shared docker network. probeDB is a sentinel
// database used only by the readiness probe (see ensureContainer).
const (
	adminPassword = "pw"
	probeDB       = "ready_probe"
)

// shared holds the one container started per test binary. started is closed once
// the container is up so Main's teardown knows whether to remove it.
var shared struct {
	once    sync.Once
	name    string
	err     error
	started bool
}

// Main runs the test binary with the shared container torn down afterward. Call
// it from a package's TestMain. If docker is unavailable the container is never
// started (individual New calls skip), and this is a plain m.Run passthrough.
func Main(m *testing.M) int {
	code := m.Run()
	if shared.started {
		_ = exec.Command("docker", "rm", "-f", shared.name).Run()
	}
	return code
}

// ensureContainer starts the shared postgres:17 container once and waits for it
// to accept queries. It skips the test (via RequireDocker) when the shared
// docker network is unavailable.
func ensureContainer(t *testing.T) string {
	t.Helper()
	harness.RequireDocker(t)
	shared.once.Do(func() {
		name := harness.UniqueName("statedb")
		args := []string{
			"run", "-d", "--network", harness.DockerNetwork(), "--name", name,
			"--tmpfs", "/var/lib/postgresql/data",
			"-e", "POSTGRES_PASSWORD=" + adminPassword,
			// A non-default POSTGRES_DB the readiness probe targets: the image's
			// temporary init server does not have it, so a successful query means
			// the real TCP server (not the init server) is up.
			"-e", "POSTGRES_DB=" + probeDB,
			"postgres:17",
		}
		if out, err := exec.Command("docker", args...).CombinedOutput(); err != nil {
			shared.err = fmt.Errorf("docker run %s: %v: %s", name, err, out)
			return
		}
		shared.name = name
		shared.started = true
		shared.err = waitReady(name)
	})
	require.NoError(t, shared.err, "start shared postgres container")
	return shared.name
}

// waitReady polls psql inside the container until the server answers a query or
// the timeout elapses (mirrors the engine suite's readiness probe: pg_isready
// reports ready too early, during the image's temporary init server).
func waitReady(name string) error {
	deadline := time.Now().Add(60 * time.Second)
	var last error
	for time.Now().Before(deadline) {
		cmd := exec.Command("docker", "exec", name,
			"psql", "-U", "postgres", "-d", probeDB, "-c", "SELECT 1")
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			last = fmt.Errorf("%v: %s", err, out)
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("postgres container %s not ready: %w", name, last)
}

// NewDSN creates a fresh, uniquely-named EMPTY database on the shared container
// and returns its connection string. The database is dropped when the test ends.
// The caller is responsible for migrating it (e.g. via db.Open, which is what
// production does) — use this when the code under test opens the database itself.
func NewDSN(t *testing.T) string {
	t.Helper()
	name := ensureContainer(t)

	dbName := "test_" + randSuffix(t)
	admin, err := sql.Open("pgx", dsn(name, "postgres"))
	require.NoError(t, err)
	defer admin.Close()
	_, err = admin.ExecContext(context.Background(), "CREATE DATABASE "+dbName)
	require.NoError(t, err, "create test database")

	t.Cleanup(func() {
		dropper, err := sql.Open("pgx", dsn(name, "postgres"))
		if err != nil {
			return
		}
		defer dropper.Close()
		// FORCE terminates any lingering connections so the drop can't hang.
		_, _ = dropper.ExecContext(context.Background(), "DROP DATABASE IF EXISTS "+dbName+" WITH (FORCE)")
	})
	return dsn(name, dbName)
}

// New creates a fresh database on the shared container (via NewDSN), migrates it
// to the current schema, and returns a handle to it. The handle is closed and
// the database dropped when the test ends.
func New(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open(NewDSN(t))
	require.NoError(t, err, "open+migrate test database")
	t.Cleanup(func() { database.Close() })
	return database
}

// dsn builds the connection string to the given database on the container,
// addressed by the container's DNS name on the shared network.
func dsn(container, dbName string) string {
	return fmt.Sprintf("postgres://postgres:%s@%s:5432/%s?sslmode=disable", adminPassword, container, dbName)
}

// randSuffix returns a random hex string safe for use in an unquoted database
// identifier (letters + digits, always starting with a letter).
func randSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 8)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return hex.EncodeToString(b)
}

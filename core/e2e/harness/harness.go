// Package harness is the shared support code for the e2e test packages
// (oauth, mcp, database). It drives the real compiled `tablekit` binary over
// HTTP: it builds the binary once per process (lazily), starts servers on random
// free ports with isolated data dirs, and offers OAuth/bearer/MCP/Docker helpers.
//
// It is a normal (non-test) package so the test packages can import it; it takes
// *testing.T and imports testing, which is fine for test-support code.
package harness

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

// moduleRoot returns core/ — three dirs up from this file (core/e2e/harness/harness.go).
func moduleRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// ensureBinary compiles the tablekit binary once per test process, to a shared
// path guarded by a cross-process file lock. `go test ./e2e/...` runs the
// package test binaries concurrently; without the lock each would launch its own
// full `go build` at the same time and exhaust memory (OOM-killed compiles). The
// lock serializes them — the first build is cold, the rest hit the build cache.
func ensureBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir := filepath.Join(os.TempDir(), "tablekit-e2e-bin")
		if buildErr = os.MkdirAll(dir, 0o755); buildErr != nil {
			return
		}
		binPath = filepath.Join(dir, "tablekit")

		lock, err := os.OpenFile(binPath+".lock", os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			buildErr = err
			return
		}
		defer lock.Close()
		if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
			buildErr = err
			return
		}
		defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

		build := exec.Command("go", "build", "-o", binPath, ".")
		build.Dir = moduleRoot()
		if out, err := build.CombinedOutput(); err != nil {
			buildErr = fmt.Errorf("go build: %w\n%s", err, out)
		}
	})
	require.NoError(t, buildErr, "building tablekit binary")
	return binPath
}

// Server is a running tablekit instance under test.
type Server struct {
	AppURL     string
	ControlURL string
	DataDir    string
	// DBEnv is the DB_* environment pointing at this server's own state database.
	// CLI helpers (RunCLI, GenerateToken) must pass it so they hit the same
	// database the server does.
	DBEnv []string
}

// StartServer starts the binary on random free ports with a fresh data dir.
func StartServer(t *testing.T) Server {
	return StartServerEnv(t)
}

// StartServerEnv is StartServer with extra "KEY=value" environment entries.
//
// The server keeps its own state in Postgres, so this provisions a throwaway
// Postgres on the shared docker network and points the binary at it — which
// means a running server needs docker; without it the test skips (RequireDocker).
func StartServerEnv(t *testing.T, extraEnv ...string) Server {
	t.Helper()
	RequireDocker(t)
	bin := ensureBinary(t)
	appPort, controlPort := freePort(t), freePort(t)
	dataDir := t.TempDir()
	appURL := fmt.Sprintf("http://127.0.0.1:%d", appPort)
	controlURL := fmt.Sprintf("http://127.0.0.1:%d", controlPort)

	cmd := exec.Command(bin, "serve")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("APP_PORT=%d", appPort),
		fmt.Sprintf("CONTROL_PORT=%d", controlPort),
		"DATA_DIR="+dataDir,
		"PUBLIC_BASE_URL="+appURL,
	)
	stateDBEnv := startStateDB(t)
	cmd.Env = append(cmd.Env, stateDBEnv...)
	cmd.Env = append(cmd.Env, extraEnv...)
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	waitHealthy(t, controlURL)
	return Server{AppURL: appURL, ControlURL: controlURL, DataDir: dataDir, DBEnv: stateDBEnv}
}

// startStateDB starts a throwaway Postgres for the server's own state and returns
// the DB_* environment the server needs to reach it. The server (a process on the
// shared docker network) reaches the database by its container name, so this
// requires docker — StartServerEnv gates on it before calling here.
func startStateDB(t *testing.T) []string {
	t.Helper()
	name := RunContainer(t, ContainerSpec{
		Name:  UniqueName("statedb"),
		Image: "postgres:17",
		Env:   []string{"POSTGRES_PASSWORD=pw", "POSTGRES_DB=tablekit"},
		Tmpfs: []string{"/var/lib/postgresql/data"},
	})
	WaitContainerReady(t, name, 60*time.Second,
		"psql", "-U", "postgres", "-d", "tablekit", "-c", "SELECT 1")
	return []string{
		"DB_HOST=" + name,
		"DB_PORT=5432",
		"DB_USER=postgres",
		"DB_PASSWORD=pw",
		"DB_NAME=tablekit",
	}
}

// freePort asks the OS for an unused TCP port, then releases it.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func waitHealthy(t *testing.T, controlURL string) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		response, err := http.Get(controlURL + "/health")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("server did not become healthy at %s", controlURL)
}

// Decode reads and JSON-unmarshals an HTTP response body into v.
func Decode(response *http.Response, v any) error {
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

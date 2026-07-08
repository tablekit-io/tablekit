package database

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"core/e2e/harness"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// dbDir returns this package's directory, which holds the container build
// contexts (containers/) and the seed dumps (test-data/).
func dbDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

// generateSSHKey returns an authorized_keys line (public) and an OpenSSH PEM
// private key for an ephemeral ed25519 pair.
func generateSSHKey(t *testing.T) (authorizedKey string, privatePEM []byte) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sshPub, err := ssh.NewPublicKey(pub)
	require.NoError(t, err)
	authorizedKey = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub)))
	block, err := ssh.MarshalPrivateKey(priv, "")
	require.NoError(t, err)
	return authorizedKey, pem.EncodeToMemory(block)
}

// startBastion builds (once) and starts the SSH bastion with the given public key.
func startBastion(t *testing.T, authorizedKey string) string {
	t.Helper()
	harness.EnsureImage(t, "tablekit-e2e-bastion:latest", filepath.Join(dbDir(), "containers", "bastion"))
	name := harness.RunContainer(t, harness.ContainerSpec{
		Name:  harness.UniqueName("bastion"),
		Image: "tablekit-e2e-bastion:latest",
		Env:   []string{"AUTHORIZED_KEY=" + authorizedKey},
	})
	harness.WaitContainerReady(t, name, 30*time.Second, "sh", "-c", "[ -f /etc/ssh/ssh_host_ed25519_key ]")
	// Give sshd a beat to bind after host keys are generated.
	time.Sleep(500 * time.Millisecond)
	return name
}

// startPostgres starts a tmpfs-backed postgres:17 seeded with cafe_data.sql and
// returns the container name (its DNS name on the shared network).
func startPostgres(t *testing.T) string {
	t.Helper()
	name := harness.RunContainer(t, harness.ContainerSpec{
		Name:  harness.UniqueName("pg"),
		Image: "postgres:17",
		Env:   []string{"POSTGRES_PASSWORD=pw", "POSTGRES_DB=cafe"},
		Tmpfs: []string{"/var/lib/postgresql/data"},
	})
	// psql (not pg_isready) as the probe: pg_isready reports ready during the
	// image's temporary init server, before POSTGRES_DB exists and the real TCP
	// server is up. A successful query against the target db means truly ready.
	harness.WaitContainerReady(t, name, 60*time.Second, "psql", "-U", "postgres", "-d", "cafe", "-c", "SELECT 1")
	seed, err := os.Open(ensureCafeSeed(t))
	require.NoError(t, err)
	defer seed.Close()
	harness.DockerExecStdin(t, name, seed, "psql", "-v", "ON_ERROR_STOP=1", "-U", "postgres", "-d", "cafe")
	return name
}

// ensureCafeSeed returns the path to cafe_data.sql, generating it first if it is
// absent. The seed is a generator artifact (gitignored), so a fresh checkout or
// CI run produces it on demand by running the committed generator.
func ensureCafeSeed(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(dbDir(), "test-data")
	seedPath := filepath.Join(dir, "cafe_data.sql")
	if _, err := os.Stat(seedPath); err == nil {
		return seedPath
	}
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "generate cafe_data.sql: %s", out)
	return seedPath
}

// startMySQL starts a tmpfs-backed mysql:8.4 seeded with dira.sql (which creates
// its own database dbctx_test_dira) and returns the container name.
func startMySQL(t *testing.T) string {
	t.Helper()
	name := harness.RunContainer(t, harness.ContainerSpec{
		Name:  harness.UniqueName("my"),
		Image: "mysql:8.4",
		Env:   []string{"MYSQL_ROOT_PASSWORD=pw"},
		Tmpfs: []string{"/var/lib/mysql"},
	})
	// Probe over TCP (not the local socket): under load the socket accepts
	// connections before mysqld is serving on TCP, which is what the engine uses.
	harness.WaitContainerReady(t, name, 90*time.Second,
		"mysql", "-uroot", "-ppw", "-h", "127.0.0.1", "-e", "SELECT 1")
	seed, err := os.Open(filepath.Join(dbDir(), "test-data", "dira.sql"))
	require.NoError(t, err)
	defer seed.Close()
	harness.DockerExecStdin(t, name, seed, "sh", "-c", "exec mysql -uroot -ppw")
	return name
}

package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The docker-backed e2e tests drive the host Docker daemon (the core container
// bind-mounts /var/run/docker.sock) and put throwaway DB/bastion containers on
// the stable `tablekit` network, which the running server also sits on, so it
// reaches them by container name. These tests only run where that network is
// available (i.e. inside the core container); elsewhere they skip.

const dockerNetworkEnv = "TABLEKIT_E2E_DOCKER_NETWORK"

var (
	dockerOnce    sync.Once
	dockerOK      bool
	imageBuildsMu sync.Mutex
	imageBuilds   = map[string]*sync.Once{}
)

// dockerAvailable reports whether docker works and the shared network is set.
func dockerAvailable() bool {
	dockerOnce.Do(func() {
		if os.Getenv(dockerNetworkEnv) == "" {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		dockerOK = exec.CommandContext(ctx, "docker", "info").Run() == nil
	})
	return dockerOK
}

// requireDocker skips the test unless docker + the shared network are available.
func requireDocker(t *testing.T) {
	t.Helper()
	if !dockerAvailable() {
		t.Skipf("docker unavailable or %s unset; skipping container e2e", dockerNetworkEnv)
	}
}

func dockerNetwork() string { return os.Getenv(dockerNetworkEnv) }

// uniqueName returns role + a random hex suffix so concurrent runs never collide.
func uniqueName(role string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return role + "-" + hex.EncodeToString(b)
}

// containerSpec describes a throwaway container to start on the shared network.
type containerSpec struct {
	name  string
	image string
	env   []string // KEY=VALUE
	tmpfs []string // in-container mount paths backed by memory
	cmd   []string // command + args appended after the image
}

// runContainer starts a detached container on the shared network and registers a
// force-remove cleanup. It returns the container name (its DNS name on the net).
func runContainer(t *testing.T, spec containerSpec) string {
	t.Helper()
	args := []string{"run", "-d", "--network", dockerNetwork(), "--name", spec.name}
	for _, mount := range spec.tmpfs {
		args = append(args, "--tmpfs", mount)
	}
	for _, env := range spec.env {
		args = append(args, "-e", env)
	}
	args = append(args, spec.image)
	args = append(args, spec.cmd...)

	out, err := exec.Command("docker", args...).CombinedOutput()
	require.NoError(t, err, "docker run %s failed: %s", spec.name, out)

	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", spec.name).Run()
	})
	return spec.name
}

// dockerExec runs a command in a container, returning combined output and error.
func dockerExec(t *testing.T, name string, argv ...string) ([]byte, error) {
	t.Helper()
	args := append([]string{"exec", name}, argv...)
	return exec.Command("docker", args...).CombinedOutput()
}

// dockerExecStdin runs a command in a container with stdin piped in (used to feed
// seed SQL to psql/mysql).
func dockerExecStdin(t *testing.T, name string, stdin io.Reader, argv ...string) {
	t.Helper()
	args := append([]string{"exec", "-i", name}, argv...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = stdin
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	require.NoError(t, cmd.Run(), "docker exec -i %s %v failed: %s", name, argv, out.String())
}

// waitContainerReady polls a readiness probe (a command run inside the container)
// until it succeeds or the timeout elapses; on timeout it dumps logs and fails.
func waitContainerReady(t *testing.T, name string, timeout time.Duration, probe ...string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := dockerExec(t, name, probe...); err == nil {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	logs, _ := exec.Command("docker", "logs", name).CombinedOutput()
	t.Fatalf("container %s not ready after %s; logs:\n%s", name, timeout, logs)
}

// ensureImage builds an image tag from a Dockerfile directory once per process.
func ensureImage(t *testing.T, tag, contextDir string) {
	t.Helper()
	imageBuildsMu.Lock()
	once, ok := imageBuilds[tag]
	if !ok {
		once = &sync.Once{}
		imageBuilds[tag] = once
	}
	imageBuildsMu.Unlock()

	var buildErr error
	var buildOut []byte
	once.Do(func() {
		buildOut, buildErr = exec.Command("docker", "build", "-t", tag, contextDir).CombinedOutput()
	})
	require.NoError(t, buildErr, "docker build %s failed: %s", tag, buildOut)
}

// e2eDir returns the absolute path to the core/e2e directory (this file's dir).
func e2eDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(moduleRoot(), "e2e")
}

package harness

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The docker-backed e2e tests drive the host Docker daemon (the core container
// bind-mounts /var/run/docker.sock) and put throwaway containers on the stable
// `tablekit` network, which the running server also sits on, so it reaches them
// by container name. These helpers only work where that network is available
// (i.e. inside the core container); elsewhere tests skip via RequireDocker.

const dockerNetworkEnv = "TABLEKIT_E2E_DOCKER_NETWORK"

var (
	dockerOnce    sync.Once
	dockerOK      bool
	imageBuildsMu sync.Mutex
	imageBuilds   = map[string]*sync.Once{}
)

// DockerAvailable reports whether docker works and the shared network is set.
func DockerAvailable() bool {
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

// RequireDocker skips the test unless docker + the shared network are available.
func RequireDocker(t *testing.T) {
	t.Helper()
	if !DockerAvailable() {
		t.Skipf("docker unavailable or %s unset; skipping container e2e", dockerNetworkEnv)
	}
}

// DockerNetwork returns the shared network name the containers attach to.
func DockerNetwork() string { return os.Getenv(dockerNetworkEnv) }

// UniqueName returns role + a random hex suffix so concurrent runs never collide.
func UniqueName(role string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return role + "-" + hex.EncodeToString(b)
}

// ContainerSpec describes a throwaway container to start on the shared network.
type ContainerSpec struct {
	Name  string
	Image string
	Env   []string // KEY=VALUE
	Tmpfs []string // in-container mount paths backed by memory
	Cmd   []string // command + args appended after the image
}

// RunContainer starts a detached container on the shared network and registers a
// force-remove cleanup. It returns the container name (its DNS name on the net).
func RunContainer(t *testing.T, spec ContainerSpec) string {
	t.Helper()
	args := []string{"run", "-d", "--network", DockerNetwork(), "--name", spec.Name}
	for _, mount := range spec.Tmpfs {
		args = append(args, "--tmpfs", mount)
	}
	for _, env := range spec.Env {
		args = append(args, "-e", env)
	}
	args = append(args, spec.Image)
	args = append(args, spec.Cmd...)

	out, err := exec.Command("docker", args...).CombinedOutput()
	require.NoError(t, err, "docker run %s failed: %s", spec.Name, out)

	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", spec.Name).Run()
	})
	return spec.Name
}

// DockerExec runs a command in a container, returning combined output and error.
func DockerExec(t *testing.T, name string, argv ...string) ([]byte, error) {
	t.Helper()
	args := append([]string{"exec", name}, argv...)
	return exec.Command("docker", args...).CombinedOutput()
}

// DockerExecStdin runs a command in a container with stdin piped in (used to feed
// seed SQL to psql/mysql).
func DockerExecStdin(t *testing.T, name string, stdin io.Reader, argv ...string) {
	t.Helper()
	args := append([]string{"exec", "-i", name}, argv...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = stdin
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	require.NoError(t, cmd.Run(), "docker exec -i %s %v failed: %s", name, argv, out.String())
}

// WaitContainerReady polls a readiness probe (a command run inside the container)
// until it succeeds or the timeout elapses; on timeout it dumps logs and fails.
func WaitContainerReady(t *testing.T, name string, timeout time.Duration, probe ...string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := DockerExec(t, name, probe...); err == nil {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	logs, _ := exec.Command("docker", "logs", name).CombinedOutput()
	t.Fatalf("container %s not ready after %s; logs:\n%s", name, timeout, logs)
}

// EnsureImage builds an image tag from a Dockerfile directory once per process.
func EnsureImage(t *testing.T, tag, contextDir string) {
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

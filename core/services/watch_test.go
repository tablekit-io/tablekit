package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"core/db/dbtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// databaseNames returns the sorted database names the engine currently exposes.
func databaseNames(t *testing.T, appServices *Services) []string {
	t.Helper()
	infos := appServices.Engine.List()
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name
	}
	return names
}

// TestWatchDatabasesFileHotReloads writes a databases file, boots the services
// over it, starts the watcher, then rewrites the file and asserts the engine
// picks up the new set without a restart.
func TestWatchDatabasesFileHotReloads(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "databases.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`
databases:
  one:
    type: postgres
    details: { host: one.internal, username: app_ro }
`), 0o600))

	t.Setenv("DATABASES_FILE", path)
	t.Setenv("DATABASE_URL", dbtest.NewDSN(t))
	t.Setenv("SIGNING_KEY", testSigningKey)

	appServices, err := New()
	require.NoError(t, err)
	t.Cleanup(func() { appServices.Close() })
	require.Equal(t, []string{"one"}, databaseNames(t, appServices))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watchDone := make(chan error, 1)
	go func() { watchDone <- appServices.WatchDatabasesFile(ctx) }()

	// Rewrite the file to a different set; the watcher should hot-reload it. The
	// write is retried on each poll (with a tick longer than the reload debounce)
	// so we don't depend on the watcher having armed before the first write, and
	// so a debounced reload always has time to settle between checks.
	next := []byte(`
databases:
  two:
    type: mysql
    details: { host: two.internal, username: reader }
`)
	require.Eventually(t, func() bool {
		require.NoError(t, os.WriteFile(path, next, 0o600))
		names := databaseNames(t, appServices)
		return len(names) == 1 && names[0] == "two"
	}, 10*time.Second, 2*reloadDebounce, "engine should reload the new databases set")

	// Cancelling the context stops the watcher cleanly.
	cancel()
	select {
	case err := <-watchDone:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop after context cancel")
	}
}

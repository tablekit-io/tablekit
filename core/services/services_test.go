package services

import (
	"os"
	"path/filepath"
	"testing"

	"core/db/dbtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available); New opens tablekit's own state database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

func TestNewWiresConfigAndStore(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("DATA_DIR", directory)
	t.Setenv("DATABASE_URL", dbtest.NewDSN(t))
	t.Setenv("SIGNING_KEY", "") // force the store-backed key path

	appServices, err := New()
	require.NoError(t, err)

	require.NotNil(t, appServices.Config)
	require.NotNil(t, appServices.Store)
	require.NotNil(t, appServices.Engine)
	// The JWT issuer is constructed once here and shared across the app.
	require.NotNil(t, appServices.Issuer)
	// Config is loaded from the environment we set above...
	assert.Equal(t, directory, appServices.Config.DataDir)
	// ...and the store is opened against that same data dir.
	_, err = appServices.Store.SigningKey()
	require.NoError(t, err)
}

func TestNewFailsOnUnusableDataDir(t *testing.T) {
	// A reachable database gets New past db.Open, so it reaches the store, which
	// can't create signing.key's directory under a regular file → New errors.
	t.Setenv("DATABASE_URL", dbtest.NewDSN(t))
	blocker := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(blocker, nil, 0o600))
	t.Setenv("DATA_DIR", filepath.Join(blocker, "sub"))

	_, err := New()
	assert.Error(t, err)
}

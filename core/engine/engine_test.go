package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeYAML writes a databases YAML to a temp file and returns its path.
func writeYAML(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "databases.yaml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func TestServiceListSortedWithType(t *testing.T) {
	path := writeYAML(t, `
databases:
  pg:
    type: postgres
    details: { host: pg.internal, username: app_ro }
  my:
    type: mysql
    details: { host: my.internal, username: reader }
`)
	svc, err := Load(path, Limits{})
	require.NoError(t, err)

	infos := svc.List()
	require.Len(t, infos, 2)
	assert.Equal(t, "my", infos[0].Name)
	assert.Equal(t, "mysql", infos[0].Type)
	assert.Equal(t, "pg", infos[1].Name)
	assert.Equal(t, "postgres", infos[1].Type)
}

func TestLimitsDefaults(t *testing.T) {
	svc, err := Load(filepath.Join(t.TempDir(), "nope.yaml"), Limits{})
	require.NoError(t, err)
	assert.Equal(t, 2048, svc.limits.MaxRows)
	assert.Equal(t, 64*1024, svc.limits.MaxBytes)
	assert.NotZero(t, svc.limits.StatementTimeout)
}

func TestLoadMissingFileTolerated(t *testing.T) {
	svc, err := Load(filepath.Join(t.TempDir(), "nope.yaml"), Limits{})
	require.NoError(t, err)
	assert.Empty(t, svc.List())
}

// names returns the sorted database names from List, so a reload's effect can be
// asserted without caring about type.
func names(infos []DatabaseInfo) []string {
	out := make([]string, len(infos))
	for i, info := range infos {
		out[i] = info.Name
	}
	return out
}

func TestReloadSwapsDatabases(t *testing.T) {
	path := writeYAML(t, `
databases:
  pg:
    type: postgres
    details: { host: pg.internal, username: app_ro }
`)
	svc, err := Load(path, Limits{})
	require.NoError(t, err)
	assert.Equal(t, []string{"pg"}, names(svc.List()))

	// Rewrite the file to a different set, then reload in place.
	require.NoError(t, os.WriteFile(path, []byte(`
databases:
  my:
    type: mysql
    details: { host: my.internal, username: reader }
`), 0o600))
	require.NoError(t, svc.Reload(path))
	assert.Equal(t, []string{"my"}, names(svc.List()))
}

func TestReloadInvalidKeepsPrevious(t *testing.T) {
	path := writeYAML(t, `
databases:
  pg:
    type: postgres
    details: { host: pg.internal, username: app_ro }
`)
	svc, err := Load(path, Limits{})
	require.NoError(t, err)

	// A malformed file must not clobber the working set.
	require.NoError(t, os.WriteFile(path, []byte("databases: [not-a-map"), 0o600))
	require.Error(t, svc.Reload(path))
	assert.Equal(t, []string{"pg"}, names(svc.List()))
}


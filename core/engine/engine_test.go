package engine

import (
	"os"
	"path/filepath"
	"testing"

	"core/engine/encoding"

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

func TestWrapPaged(t *testing.T) {
	assert.Equal(t,
		"SELECT * FROM (SELECT id FROM t) AS _tablekit_page LIMIT 11 OFFSET 0",
		wrapPaged("SELECT id FROM t", 0, 10))

	// Trailing semicolon and surrounding whitespace are stripped so the query
	// is a valid subquery; offset and limit+1 over-fetch are applied.
	assert.Equal(t,
		"SELECT * FROM (SELECT id FROM t) AS _tablekit_page LIMIT 6 OFFSET 20",
		wrapPaged("  SELECT id FROM t ;  ", 20, 5))
}

func TestTrimToLimit(t *testing.T) {
	// One extra row over the limit -> trimmed back to limit, hasMore true.
	over := &encoding.Result{
		Rows:     []map[string]any{{"n": 1}, {"n": 2}, {"n": 3}},
		RowCount: 3,
	}
	hasMore := trimToLimit(over, 2)
	assert.True(t, hasMore)
	assert.Equal(t, 2, over.RowCount)
	assert.Len(t, over.Rows, 2)

	// Exactly the limit (no over-fetched row) -> untouched, hasMore false.
	exact := &encoding.Result{
		Rows:     []map[string]any{{"n": 1}, {"n": 2}},
		RowCount: 2,
	}
	hasMore = trimToLimit(exact, 2)
	assert.False(t, hasMore)
	assert.Equal(t, 2, exact.RowCount)
	assert.Len(t, exact.Rows, 2)
}

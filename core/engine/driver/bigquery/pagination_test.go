package bigquery

import (
	"math"
	"testing"

	"core/engine/config"
	"core/engine/encoding"

	"github.com/stretchr/testify/assert"
)

func TestPageQuery(t *testing.T) {
	// Window applied: wrapped in a subquery, over-fetching one row (limit+1).
	assert.Equal(t,
		"SELECT * FROM (SELECT id FROM t) AS _tablekit_page LIMIT 11 OFFSET 0",
		pageQuery("SELECT id FROM t", config.Page{Offset: 0, Limit: 10}))

	// Trailing semicolon and surrounding whitespace stripped so the query is a
	// valid subquery; offset and limit+1 over-fetch applied.
	assert.Equal(t,
		"SELECT * FROM (SELECT id FROM t) AS _tablekit_page LIMIT 6 OFFSET 20",
		pageQuery("  SELECT id FROM t ;  ", config.Page{Offset: 20, Limit: 5}))

	// No window (Limit <= 0): query returned unwrapped.
	assert.Equal(t, "SELECT id FROM t",
		pageQuery("SELECT id FROM t", config.Page{}))
}

func TestAccumulatorRows(t *testing.T) {
	assert.Equal(t, 11, accumulatorRows(config.Page{Limit: 10}))
	assert.Equal(t, math.MaxInt, accumulatorRows(config.Page{Limit: 0}))
}

func TestTrimToLimit(t *testing.T) {
	over := &encoding.Result{
		Rows:     []map[string]any{{"n": 1}, {"n": 2}, {"n": 3}},
		RowCount: 3,
	}
	assert.True(t, trimToLimit(over, 2))
	assert.Equal(t, 2, over.RowCount)
	assert.Len(t, over.Rows, 2)

	exact := &encoding.Result{
		Rows:     []map[string]any{{"n": 1}, {"n": 2}},
		RowCount: 2,
	}
	assert.False(t, trimToLimit(exact, 2))
	assert.Equal(t, 2, exact.RowCount)

	unbounded := &encoding.Result{
		Rows:     []map[string]any{{"n": 1}, {"n": 2}, {"n": 3}},
		RowCount: 3,
	}
	assert.False(t, trimToLimit(unbounded, 0))
	assert.Equal(t, 3, unbounded.RowCount)
}

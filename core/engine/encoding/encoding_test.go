package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccumulatorRowCap(t *testing.T) {
	acc := NewAccumulator([]string{"n"}, 2, 1<<20)
	for i := 0; i < 5; i++ {
		keep, err := acc.Add(map[string]any{"n": i})
		require.NoError(t, err)
		if i < 2 {
			assert.True(t, keep, "row %d should be kept", i)
		} else {
			assert.False(t, keep, "row %d should hit the cap", i)
			break
		}
	}
	result := acc.Result()
	assert.Equal(t, 2, result.RowCount)
	assert.True(t, result.Truncated)
}

func TestAccumulatorByteCap(t *testing.T) {
	// Cap small enough that the second row overflows, but the first is always kept.
	acc := NewAccumulator([]string{"v"}, 100, 20)
	keep, err := acc.Add(map[string]any{"v": "first-row-is-big-enough"})
	require.NoError(t, err)
	assert.True(t, keep, "first row always kept even if over the byte cap")

	keep, err = acc.Add(map[string]any{"v": "second"})
	require.NoError(t, err)
	assert.False(t, keep, "second row should overflow the byte cap")

	result := acc.Result()
	assert.Equal(t, 1, result.RowCount)
	assert.True(t, result.Truncated)
}

func TestAccumulatorOmitDedupeAndSort(t *testing.T) {
	acc := NewAccumulator([]string{"a", "b"}, 10, 1<<20)
	acc.Omit("zeta", "reason-z")
	acc.Omit("alpha", "reason-a")
	acc.Omit("zeta", "reason-z-again") // dedupe: first reason wins

	result := acc.Result()
	require.Len(t, result.Omitted, 2)
	assert.Equal(t, "alpha", result.Omitted[0].Column)
	assert.Equal(t, "zeta", result.Omitted[1].Column)
	assert.Equal(t, "reason-z", result.Omitted[1].Reason)
}

func TestAccumulatorEmptyRowsNonNil(t *testing.T) {
	acc := NewAccumulator([]string{"a"}, 10, 1<<20)
	result := acc.Result()
	assert.NotNil(t, result.Rows)
	assert.Equal(t, 0, result.RowCount)
	assert.False(t, result.Truncated)
	assert.Empty(t, result.Omitted)
}

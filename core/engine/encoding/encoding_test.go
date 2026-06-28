package encoding

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeNumeric stands in for a pgtype.* value: it implements driver.Valuer and
// unwraps to a string, the way pgx returns numerics to preserve precision.
type fakeNumeric struct{ s string }

func (f fakeNumeric) Value() (driver.Value, error) { return f.s, nil }

func TestNormalizeValue(t *testing.T) {
	when := time.Date(2026, 6, 28, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name   string
		in     any
		want   any
		keep   bool
		reason string
	}{
		{"nil", nil, nil, true, ""},
		{"time", when, "2026-06-28T12:30:00Z", true, ""},
		{"utf8 bytes", []byte("héllo"), "héllo", true, ""},
		{"non-utf8 bytes", []byte{0xff, 0xfe, 0x00}, nil, false, binaryOmitReason},
		{"int64", int64(42), int64(42), true, ""},
		{"float64", 3.5, 3.5, true, ""},
		{"bool", true, true, true, ""},
		{"string", "plain", "plain", true, ""},
		{"valuer numeric", fakeNumeric{"123.456"}, "123.456", true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, keep, reason := NormalizeValue(tt.in)
			assert.Equal(t, tt.keep, keep)
			assert.Equal(t, tt.reason, reason)
			if tt.keep {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

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

package postgres

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
			got, keep, reason := normalizeValue(tt.in)
			assert.Equal(t, tt.keep, keep)
			assert.Equal(t, tt.reason, reason)
			if tt.keep {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

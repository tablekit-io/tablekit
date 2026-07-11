package bigquery

import (
	"math/big"
	"testing"
	"time"

	"cloud.google.com/go/civil"

	bigqueryapi "cloud.google.com/go/bigquery"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeValueScalars(t *testing.T) {
	when := time.Date(2026, 6, 28, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name  string
		in    any
		field *bigqueryapi.FieldSchema
		want  any
		keep  bool
	}{
		{"nil", nil, nil, nil, true},
		{"bool", true, nil, true, true},
		{"string", "plain", nil, "plain", true},
		{"int64", int64(42), nil, int64(42), true},
		{"float64", 3.5, nil, 3.5, true},
		{"bytes base64", []byte("hi"), &bigqueryapi.FieldSchema{Type: bigqueryapi.BytesFieldType}, "aGk=", true},
		{"timestamp", when, nil, "2026-06-28T12:30:00Z", true},
		{"date", civil.Date{Year: 2026, Month: 6, Day: 28}, nil, "2026-06-28", true},
		{"datetime", civil.DateTime{Date: civil.Date{Year: 2026, Month: 6, Day: 28}, Time: civil.Time{Hour: 1}}, nil, "2026-06-28T01:00:00", true},
		{"unknown omitted", struct{ X int }{1}, nil, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, keep, reason := normalizeValue(tt.in, tt.field)
			assert.Equal(t, tt.keep, keep)
			if tt.keep {
				assert.Equal(t, tt.want, got)
			} else {
				assert.Equal(t, binaryOmitReason, reason)
			}
		})
	}
}

// NUMERIC and BIGNUMERIC both arrive as *big.Rat; the field type selects the
// formatter, so a value formats differently under each.
func TestNormalizeValueNumeric(t *testing.T) {
	rat := big.NewRat(3, 2)

	numeric, keep, _ := normalizeValue(rat, &bigqueryapi.FieldSchema{Type: bigqueryapi.NumericFieldType})
	assert.True(t, keep)
	assert.Equal(t, bigqueryapi.NumericString(rat), numeric)

	bigNumeric, keep, _ := normalizeValue(rat, &bigqueryapi.FieldSchema{Type: bigqueryapi.BigNumericFieldType})
	assert.True(t, keep)
	assert.Equal(t, bigqueryapi.BigNumericString(rat), bigNumeric)
}

// A repeated field arrives as a list; each element is normalized as a scalar of
// the same field.
func TestNormalizeValueRepeated(t *testing.T) {
	field := &bigqueryapi.FieldSchema{Type: bigqueryapi.StringFieldType, Repeated: true}
	got, keep, _ := normalizeValue([]bigqueryapi.Value{"a", "b"}, field)
	assert.True(t, keep)
	assert.Equal(t, []any{"a", "b"}, got)
}

// A RECORD arrives as one value per nested field, in schema order; the result is
// a map keyed by the nested field names, recursively normalized.
func TestNormalizeValueRecord(t *testing.T) {
	field := &bigqueryapi.FieldSchema{
		Type: bigqueryapi.RecordFieldType,
		Schema: bigqueryapi.Schema{
			{Name: "id", Type: bigqueryapi.IntegerFieldType},
			{Name: "tags", Type: bigqueryapi.StringFieldType, Repeated: true},
		},
	}
	got, keep, _ := normalizeValue([]bigqueryapi.Value{int64(7), []bigqueryapi.Value{"x", "y"}}, field)
	assert.True(t, keep)
	assert.Equal(t, map[string]any{"id": int64(7), "tags": []any{"x", "y"}}, got)
}

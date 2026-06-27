package engine

import (
	"database/sql/driver"
	"encoding/json"
	"sort"
	"time"
	"unicode/utf8"
)

// binaryOmitReason is recorded for values that cannot be represented as JSON
// text. It is surfaced to the caller (and the LLM) via Result.Omitted.
const binaryOmitReason = "binary (non-UTF-8) value not representable as JSON text"

// rowAccumulator collects normalized rows under the configured limits, tracking
// truncation and any columns that had to be dropped.
type rowAccumulator struct {
	limits    Limits
	columns   []string
	rows      []map[string]any
	bytes     int
	truncated bool
	omitted   map[string]string // column name -> reason
}

func newRowAccumulator(columns []string, limits Limits) *rowAccumulator {
	return &rowAccumulator{
		limits:  limits,
		columns: columns,
		omitted: map[string]string{},
	}
}

// add appends a normalized row. It returns false when a row or byte cap is hit,
// signalling the caller to stop reading; truncated is set in that case.
func (a *rowAccumulator) add(row map[string]any) (bool, error) {
	if len(a.rows) >= a.limits.MaxRows {
		a.truncated = true
		return false, nil
	}
	encoded, err := json.Marshal(row)
	if err != nil {
		return false, err
	}
	// Always keep at least one row, even if it alone exceeds the byte cap.
	if len(a.rows) > 0 && a.bytes+len(encoded) > a.limits.MaxBytes {
		a.truncated = true
		return false, nil
	}
	a.bytes += len(encoded)
	a.rows = append(a.rows, row)
	return true, nil
}

// omit records that a column was dropped, keeping the first reason seen.
func (a *rowAccumulator) omit(column, reason string) {
	if _, seen := a.omitted[column]; !seen {
		a.omitted[column] = reason
	}
}

func (a *rowAccumulator) result() *Result {
	omitted := make([]OmittedColumn, 0, len(a.omitted))
	for column, reason := range a.omitted {
		omitted = append(omitted, OmittedColumn{Column: column, Reason: reason})
	}
	sort.Slice(omitted, func(i, j int) bool { return omitted[i].Column < omitted[j].Column })

	rows := a.rows
	if rows == nil {
		rows = []map[string]any{}
	}
	return &Result{
		Columns:   a.columns,
		Rows:      rows,
		RowCount:  len(a.rows),
		Truncated: a.truncated,
		Omitted:   omitted,
	}
}

// normalizeValue converts a driver-decoded value into a JSON-friendly one. The
// bool is false when the value must be omitted (and the string gives the
// reason). It handles the value shapes both pgx (rows.Values) and database/sql
// (scan into any) produce.
func normalizeValue(value any) (any, bool, string) {
	switch typed := value.(type) {
	case nil:
		return nil, true, ""
	case time.Time:
		return typed.Format(time.RFC3339Nano), true, ""
	case []byte:
		if utf8.Valid(typed) {
			return string(typed), true, ""
		}
		return nil, false, binaryOmitReason
	case bool, string,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return typed, true, ""
	default:
		// pgtype.* values (numeric, uuid, ...) implement driver.Valuer; unwrap
		// to the underlying basic value (numerics come back as strings, which
		// preserves precision) and normalize that.
		if valuer, ok := value.(driver.Valuer); ok {
			underlying, err := valuer.Value()
			if err == nil && underlying != value {
				return normalizeValue(underlying)
			}
		}
		return typed, true, ""
	}
}

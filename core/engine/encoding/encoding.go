// Package encoding turns driver-decoded rows into JSON-friendly results under
// caller-supplied caps. It owns the public Result shape and the value
// normalization rules, and depends on nothing else in engine so drivers can use
// it without dragging in config or connection concerns.
package encoding

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

// Result is a query's outcome: the column names, the rows as objects, and flags
// telling the caller when the result was cut short or had columns dropped.
type Result struct {
	Columns   []string         `json:"columns"`
	Rows      []map[string]any `json:"rows"`
	RowCount  int              `json:"row_count"`
	Truncated bool             `json:"truncated"`
	// Omitted lists columns whose values could not be represented as JSON text
	// (e.g. binary/non-UTF-8) and were dropped from Rows, so the caller knows
	// the column exists but was excluded, and why.
	Omitted []OmittedColumn `json:"omitted,omitempty"`
}

// OmittedColumn names a dropped column and the reason it was dropped.
type OmittedColumn struct {
	Column string `json:"column"`
	Reason string `json:"reason"`
}

// Accumulator collects normalized rows under the configured caps, tracking
// truncation and any columns that had to be dropped.
type Accumulator struct {
	columns   []string
	maxRows   int
	maxBytes  int
	rows      []map[string]any
	bytes     int
	truncated bool
	omitted   map[string]string // column name -> reason
}

// NewAccumulator builds an accumulator that keeps at most maxRows rows and
// maxBytes of encoded row data (whichever cap is hit first truncates).
func NewAccumulator(columns []string, maxRows, maxBytes int) *Accumulator {
	return &Accumulator{
		columns:  columns,
		maxRows:  maxRows,
		maxBytes: maxBytes,
		omitted:  map[string]string{},
	}
}

// Add appends a normalized row. It returns false when a row or byte cap is hit,
// signalling the caller to stop reading; truncated is set in that case.
func (a *Accumulator) Add(row map[string]any) (bool, error) {
	if len(a.rows) >= a.maxRows {
		a.truncated = true
		return false, nil
	}
	encoded, err := json.Marshal(row)
	if err != nil {
		return false, err
	}
	// Always keep at least one row, even if it alone exceeds the byte cap.
	if len(a.rows) > 0 && a.bytes+len(encoded) > a.maxBytes {
		a.truncated = true
		return false, nil
	}
	a.bytes += len(encoded)
	a.rows = append(a.rows, row)
	return true, nil
}

// Omit records that a column was dropped, keeping the first reason seen.
func (a *Accumulator) Omit(column, reason string) {
	if _, seen := a.omitted[column]; !seen {
		a.omitted[column] = reason
	}
}

// Truncated reports whether a cap has been hit so far.
func (a *Accumulator) Truncated() bool { return a.truncated }

// Result snapshots the accumulated rows into a Result.
func (a *Accumulator) Result() *Result {
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

// NormalizeValue converts a driver-decoded value into a JSON-friendly one. The
// bool is false when the value must be omitted (and the string gives the
// reason). It handles the value shapes both pgx (rows.Values) and database/sql
// (scan into any) produce.
func NormalizeValue(value any) (any, bool, string) {
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
				return NormalizeValue(underlying)
			}
		}
		return typed, true, ""
	}
}

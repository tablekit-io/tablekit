// Package encoding turns driver-decoded rows into JSON-friendly results under
// caller-supplied caps. It owns the public Result shape and the Accumulator that
// enforces the caps, and depends on nothing else in engine so drivers can use it
// without dragging in config or connection concerns. Value normalization is not
// here: each driver owns its own, since the value shapes differ per driver.
package encoding

import (
	"encoding/json"
	"sort"
)

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

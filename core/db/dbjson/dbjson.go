// Package dbjson provides a generic wrapper for a jsonb column whose shape is
// known, so generated go-jet model fields can be strongly typed instead of raw
// bytes. It is a leaf package (stdlib only) so the generated model package may
// import it without a cycle.
package dbjson

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSON wraps a value of type T stored in a jsonb column. It implements
// sql.Scanner and driver.Valuer, so database/sql marshals the value to JSON on
// write and unmarshals it on read automatically.
type JSON[T any] struct {
	Val T
}

// Value marshals the wrapped value to JSON for the driver.
func (j JSON[T]) Value() (driver.Value, error) {
	return json.Marshal(j.Val)
}

// Scan unmarshals jsonb bytes (or text) from the driver into the wrapped value.
// A SQL NULL leaves the wrapped value at its zero.
func (j *JSON[T]) Scan(src any) error {
	if src == nil {
		var zero T
		j.Val = zero
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("dbjson: cannot scan %T into JSON", src)
	}
	return json.Unmarshal(data, &j.Val)
}

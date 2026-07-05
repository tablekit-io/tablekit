package mysql

import (
	"database/sql/driver"
	"time"
	"unicode/utf8"
)

// binaryOmitReason is recorded for values that cannot be represented as JSON
// text. It is surfaced to the caller (and the LLM) via Result.Omitted.
const binaryOmitReason = "binary (non-UTF-8) value not representable as JSON text"

// normalizeValue converts a database/sql-decoded value into a JSON-friendly one.
// The bool is false when the value must be omitted (and the string gives the
// reason). It handles the value shapes go-sql-driver's scan-into-any produces:
// []byte for most text/blob columns, int64/float64 for numerics, time.Time for
// dates (parseTime), and bool.
//
// Each driver owns its own copy of this so it can normalize exactly the types its
// driver library yields, without a shared switch having to know every driver's
// value shapes. The Postgres copy is intentionally similar today but free to
// diverge (e.g. pgtype.* unwrapping matters there, not here).
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
		// Some drivers wrap values behind driver.Valuer; unwrap to the underlying
		// basic value and normalize that.
		if valuer, ok := value.(driver.Valuer); ok {
			underlying, err := valuer.Value()
			if err == nil && underlying != value {
				return normalizeValue(underlying)
			}
		}
		return typed, true, ""
	}
}

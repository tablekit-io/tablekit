package postgres

import (
	"database/sql/driver"
	"time"
	"unicode/utf8"
)

// binaryOmitReason is recorded for values that cannot be represented as JSON
// text. It is surfaced to the caller (and the LLM) via Result.Omitted.
const binaryOmitReason = "binary (non-UTF-8) value not representable as JSON text"

// normalizeValue converts a pgx-decoded value into a JSON-friendly one. The bool
// is false when the value must be omitted (and the string gives the reason). It
// handles the value shapes pgx's rows.Values produces: basic Go types, time.Time,
// []byte, and pgtype.* values that implement driver.Valuer.
//
// Each driver owns its own copy of this so it can normalize exactly the types its
// driver library yields, without a shared switch having to know every driver's
// value shapes. The MySQL copy is intentionally similar today but free to diverge.
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

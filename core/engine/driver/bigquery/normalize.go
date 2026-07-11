package bigquery

import (
	"encoding/base64"
	"math/big"
	"time"

	"cloud.google.com/go/civil"

	bigqueryapi "cloud.google.com/go/bigquery"
)

// binaryOmitReason mirrors the postgres driver: a value that cannot be
// represented as JSON text is dropped and this reason surfaced to the caller.
const binaryOmitReason = "value not representable as JSON text"

// normalizeValue converts a BigQuery-decoded value into a JSON-friendly one. The
// bool is false when the value must be omitted (and the string gives the reason).
//
// Unlike the SQL drivers it is schema-aware: BigQuery hands nested STRUCT/ARRAY
// values back as untyped []bigquery.Value lists, and NUMERIC vs BIGNUMERIC both
// arrive as *big.Rat, so the field schema is needed to recover nested field names
// and pick the right decimal formatting. field may be nil for a value with no
// schema, in which case only the Go value type drives the conversion.
func normalizeValue(value any, field *bigqueryapi.FieldSchema) (any, bool, string) {
	if value == nil {
		return nil, true, ""
	}

	// Repeated fields arrive as a list of element values; normalize each as a
	// non-repeated element of the same field. A non-representable element is
	// dropped from the array rather than dropping the whole column.
	if field != nil && field.Repeated {
		elements, ok := value.([]bigqueryapi.Value)
		if !ok {
			return nil, false, binaryOmitReason
		}
		element := *field
		element.Repeated = false
		out := make([]any, 0, len(elements))
		for _, raw := range elements {
			normalized, keep, _ := normalizeValue(raw, &element)
			if keep {
				out = append(out, normalized)
			}
		}
		return out, true, ""
	}

	// RECORD (STRUCT) arrives as one value per nested field, in schema order.
	if field != nil && field.Type == bigqueryapi.RecordFieldType {
		return normalizeRecord(value, field.Schema)
	}

	switch typed := value.(type) {
	case bool, string,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return typed, true, ""
	case *big.Rat:
		if field != nil && field.Type == bigqueryapi.BigNumericFieldType {
			return bigqueryapi.BigNumericString(typed), true, ""
		}
		return bigqueryapi.NumericString(typed), true, ""
	case []byte:
		return base64.StdEncoding.EncodeToString(typed), true, ""
	case time.Time:
		return typed.Format(time.RFC3339Nano), true, ""
	case civil.Date:
		return typed.String(), true, ""
	case civil.Time:
		return typed.String(), true, ""
	case civil.DateTime:
		return typed.String(), true, ""
	default:
		return nil, false, binaryOmitReason
	}
}

// normalizeRecord turns a STRUCT value (a list of nested field values in schema
// order) into a map keyed by the nested field names. A nested field whose value
// cannot be represented is left out of the map.
func normalizeRecord(value any, schema bigqueryapi.Schema) (any, bool, string) {
	elements, ok := value.([]bigqueryapi.Value)
	if !ok {
		return nil, false, binaryOmitReason
	}
	record := make(map[string]any, len(schema))
	for index, nested := range schema {
		if index >= len(elements) {
			break
		}
		normalized, keep, _ := normalizeValue(elements[index], nested)
		if keep {
			record[nested.Name] = normalized
		}
	}
	return record, true, ""
}

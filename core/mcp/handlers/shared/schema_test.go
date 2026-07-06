package shared

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectColumnsEmptyRequestPassesThrough(t *testing.T) {
	columns := []string{"a", "b"}
	rows := []map[string]any{{"a": 1, "b": 2}}

	gotCols, gotRows := ProjectColumns(columns, rows, nil)
	assert.Equal(t, columns, gotCols)
	assert.Equal(t, rows, gotRows)
}

func TestProjectColumnsSelectsAndReorders(t *testing.T) {
	columns := []string{"a", "b", "c"}
	rows := []map[string]any{
		{"a": 1, "b": 2, "c": 3},
		{"a": 4, "b": 5, "c": 6},
	}

	// Request a reordered subset; unknown names are dropped silently.
	gotCols, gotRows := ProjectColumns(columns, rows, []string{"c", "a", "missing"})
	assert.Equal(t, []string{"c", "a"}, gotCols)
	assert.Equal(t, []map[string]any{
		{"c": 3, "a": 1},
		{"c": 6, "a": 4},
	}, gotRows)
}

// compatType is the struct the schema-guard tests reflect against. It exercises
// a scalar, an optional scalar, a nested object and an array of objects.
type compatType struct {
	Name   string `json:"name"`
	Count  int    `json:"count,omitempty"`
	Nested struct {
		Prop string `json:"prop"`
	} `json:"nested"`
	Items []struct {
		Kind string `json:"kind"`
	} `json:"items"`
}

func TestAssertCompatibleAcceptsMatchingSchemaWithRefinements(t *testing.T) {
	// A hand-written schema that matches the struct shape and adds an enum the
	// struct can't express — the guard must accept the refinement.
	var schema jsonschema.Schema
	require.NoError(t, schema.UnmarshalJSON([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "enum": ["a", "b"]},
			"count": {"type": "integer"},
			"nested": {"type": "object", "properties": {"prop": {"type": "string"}}},
			"items": {"type": "array", "items": {"type": "object", "properties": {"kind": {"type": "string"}}}}
		}
	}`)))
	assert.NoError(t, assertCompatible[compatType](&schema))
}

func TestAssertCompatibleRejectsUnknownProperty(t *testing.T) {
	var schema jsonschema.Schema
	require.NoError(t, schema.UnmarshalJSON([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"count": {"type": "integer"},
			"nested": {"type": "object", "properties": {"prop": {"type": "string"}}},
			"items": {"type": "array", "items": {"type": "object", "properties": {"kind": {"type": "string"}}}},
			"stray": {"type": "string"}
		}
	}`)))
	assert.Error(t, assertCompatible[compatType](&schema))
}

func TestAssertCompatibleRejectsTypeMismatch(t *testing.T) {
	var schema jsonschema.Schema
	require.NoError(t, schema.UnmarshalJSON([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "integer"},
			"count": {"type": "integer"},
			"nested": {"type": "object", "properties": {"prop": {"type": "string"}}},
			"items": {"type": "array", "items": {"type": "object", "properties": {"kind": {"type": "string"}}}}
		}
	}`)))
	assert.Error(t, assertCompatible[compatType](&schema))
}

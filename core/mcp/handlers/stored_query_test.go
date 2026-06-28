package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectColumnsEmptyRequestPassesThrough(t *testing.T) {
	columns := []string{"a", "b"}
	rows := []map[string]any{{"a": 1, "b": 2}}

	gotCols, gotRows := projectColumns(columns, rows, nil)
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
	gotCols, gotRows := projectColumns(columns, rows, []string{"c", "a", "missing"})
	assert.Equal(t, []string{"c", "a"}, gotCols)
	assert.Equal(t, []map[string]any{
		{"c": 3, "a": 1},
		{"c": 6, "a": 4},
	}, gotRows)
}

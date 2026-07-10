package shared

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMustTemplatePanicsOnParseError locks in the fail-at-startup contract: a
// malformed template must panic when the tool package initializes, never at
// request time.
func TestMustTemplatePanicsOnParseError(t *testing.T) {
	assert.Panics(t, func() { MustTemplate([]byte(`{% for x in %}`)) })
}

// TestRenderTextIsLossless exercises the whole canonical-output → text path for
// the two things that make a text rendering faithful: exact numeric values and
// unescaped, unmangled cell content.
func TestRenderTextIsLossless(t *testing.T) {
	tmpl := MustTemplate([]byte(`{% autoescape off %}{{ rows|mdtable:columns }}{% endautoescape %}`))

	out := struct {
		Columns []string         `json:"columns"`
		Rows    []map[string]any `json:"rows"`
	}{
		Columns: []string{"id", "note"},
		Rows: []map[string]any{
			// 9007199254740993 = 2^53+1, unrepresentable as float64 — proves the
			// UseNumber round-trip keeps big ints exact rather than flattening them.
			{"id": int64(9007199254740993), "note": "a|b"},
			{"id": nil, "note": "line1\nline2"},
			{"id": 3, "note": "<b> & </b>"},
		},
	}

	text, err := RenderText(tmpl, out)
	require.NoError(t, err)

	assert.Contains(t, text, "| id | note |")
	assert.Contains(t, text, "| --- | --- |")
	assert.Contains(t, text, "9007199254740993", "big int must survive losslessly")
	assert.Contains(t, text, "a\\|b", "pipe must be escaped so it can't break the column")
	assert.Contains(t, text, "line1<br>line2", "newline becomes <br> to stay on the row")
	assert.Contains(t, text, "<b> & </b>", "html chars must not be entity-escaped")
	// nil renders as an empty cell, not the literal "<nil>".
	assert.NotContains(t, text, "<nil>")
}

// TestMarkdownTableAcceptsObjectColumns covers the query_database column shape
// (ColumnInfo objects, `{"name": ...}`) as opposed to read_results' bare
// strings — the mdtable filter must resolve both.
func TestMarkdownTableAcceptsObjectColumns(t *testing.T) {
	tmpl := MustTemplate([]byte(`{% autoescape off %}{{ rows|mdtable:columns }}{% endautoescape %}`))

	out := struct {
		Columns []ColumnInfo     `json:"columns"`
		Rows    []map[string]any `json:"rows"`
	}{
		Columns: []ColumnInfo{{Name: "city"}, {Name: "total"}},
		Rows:    []map[string]any{{"city": "Dhaka", "total": 42}},
	}

	text, err := RenderText(tmpl, out)
	require.NoError(t, err)
	lines := strings.Split(text, "\n")
	require.Len(t, lines, 3, "header, separator, one data row")
	assert.Equal(t, "| city | total |", lines[0])
	assert.Equal(t, "| --- | --- |", lines[1])
	assert.Equal(t, "| Dhaka | 42 |", lines[2])
}

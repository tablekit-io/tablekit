package querydatabase

import (
	"testing"

	"core/mcp/handlers/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputTemplateRendersInlinedRows renders the real embedded output.tmpl
// against a populated output so a field-name typo in the template fails here
// rather than silently dropping data from the text content clients show Claude.
func TestOutputTemplateRendersInlinedRows(t *testing.T) {
	out := output{
		Database:     "cafe",
		ResultKey:    "abc-123",
		RowCount:     2,
		HasMore:      true,
		DefaultLimit: shared.DefaultLimit,
		Columns:      []shared.ColumnInfo{{Name: "city"}, {Name: "orders"}},
		Rows: []map[string]any{
			{"city": "Dhaka", "orders": 5},
			{"city": "New York", "orders": 3},
		},
		HintsForAIAgents: []string{shared.PaginationHint, shared.ChartHint},
	}

	text, err := shared.RenderText(textTemplate, out)
	require.NoError(t, err)

	assert.Contains(t, text, "## Query Results")
	assert.Contains(t, text, "Stored query `abc-123` against **cafe** — 2 row(s) in the first page (more rows available).")
	assert.Contains(t, text, "| city | orders |")
	assert.Contains(t, text, "| --- | --- |")
	assert.Contains(t, text, "| Dhaka | 5 |")
	assert.Contains(t, text, "| New York | 3 |")
	assert.Contains(t, text, "**Hints for AI agents**")
	assert.Contains(t, text, shared.PaginationHint)
	assert.Contains(t, text, shared.ChartHint)
}

// TestOutputTemplateOmitsTableWithoutRows covers the no-inline-results path: the
// summary and next-step guidance render, but no empty table is emitted.
func TestOutputTemplateOmitsTableWithoutRows(t *testing.T) {
	out := output{
		Database:     "cafe",
		ResultKey:    "abc-123",
		RowCount:     42,
		DefaultLimit: shared.DefaultLimit,
		Columns:      []shared.ColumnInfo{{Name: "city"}},
	}

	text, err := shared.RenderText(textTemplate, out)
	require.NoError(t, err)

	assert.Contains(t, text, "42 row(s) in the first page.")
	assert.NotContains(t, text, "| --- |")
	assert.NotContains(t, text, shared.ChartHint)
}

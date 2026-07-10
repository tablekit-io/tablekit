package readresults

import (
	"testing"

	"core/mcp/handlers/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputTemplateRendersWindow renders the real embedded output.tmpl against
// a populated window so a template field-name typo fails here rather than
// silently dropping rows from the text content.
func TestOutputTemplateRendersWindow(t *testing.T) {
	next := 128
	out := output{
		Key:              "abc-123",
		Skip:             0,
		Limit:            128,
		Columns:          []string{"city", "orders"},
		Rows:             []map[string]any{{"city": "Dhaka", "orders": 5}},
		RowsReturned:     1,
		HasMore:          true,
		NextSkip:         &next,
		HintsForAIAgents: []string{shared.ChartHint},
	}

	text, err := shared.RenderText(textTemplate, out)
	require.NoError(t, err)

	assert.Contains(t, text, "Window at skip 0 (limit 128) of stored query abc-123: 1 row(s) (more rows available, next_skip 128).")
	assert.Contains(t, text, "| city | orders |")
	assert.Contains(t, text, "| Dhaka | 5 |")
	assert.Contains(t, text, shared.ChartHint)
}

package listavailabledatabases

import (
	"testing"

	"core/engine"
	"core/mcp/handlers/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputTemplateRendersDatabaseList renders the real embedded output.tmpl so a
// field-name typo in the template fails here rather than dropping databases from the
// text content. It also pins the human-friendly markdown shape (heading + bullet list).
func TestOutputTemplateRendersDatabaseList(t *testing.T) {
	out := output{
		Databases: []engine.DatabaseInfo{
			{Name: "cafe", Type: "postgres"},
			{Name: "warehouse", Type: "bigquery"},
		},
		HintsForAIAgents: []string{bigQueryCostHint},
	}

	text, err := shared.RenderText(textTemplate, out)
	require.NoError(t, err)

	assert.Contains(t, text, "## Available Databases")
	assert.Contains(t, text, "2 database(s) configured.")
	assert.Contains(t, text, "- **cafe** (`postgres`)")
	assert.Contains(t, text, "- **warehouse** (`bigquery`)")
	assert.Contains(t, text, "**Hints for AI agents**")
	assert.Contains(t, text, bigQueryCostHint)
}

// TestOutputTemplateOmitsHintsWhenNone covers the no-hints path (no BigQuery database):
// the database list renders but no empty "Hints for AI agents" section is emitted.
func TestOutputTemplateOmitsHintsWhenNone(t *testing.T) {
	out := output{
		Databases: []engine.DatabaseInfo{{Name: "cafe", Type: "postgres"}},
	}

	text, err := shared.RenderText(textTemplate, out)
	require.NoError(t, err)

	assert.Contains(t, text, "- **cafe** (`postgres`)")
	assert.NotContains(t, text, "**Hints for AI agents**")
}

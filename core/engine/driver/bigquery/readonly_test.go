package bigquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkStatementType is the authoritative read-only decision. The dry-run
// plumbing that feeds it a real StatementType is exercised by the emulator e2e;
// here we pin the allow/deny rule itself.
func TestCheckStatementType(t *testing.T) {
	require.NoError(t, checkStatementType("SELECT"), "a plain SELECT is allowed")

	for _, statementType := range []string{"INSERT", "UPDATE", "DELETE", "MERGE", "CREATE_TABLE", "DROP_TABLE", "SCRIPT", ""} {
		err := checkStatementType(statementType)
		require.Error(t, err, "%q must be refused", statementType)
		assert.Contains(t, err.Error(), statementType)
	}
}

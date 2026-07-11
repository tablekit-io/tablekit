package bigquery

import (
	"context"
	"strings"
	"testing"

	"core/engine/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bigQueryConfig(projectID, location string) config.Database {
	return config.Database{
		Name:     "target",
		Type:     config.DatabaseTypeBigQuery,
		BigQuery: &config.BigQueryDetails{ProjectID: projectID, Location: location},
	}
}

// DeriveIdentity is config-derived (project ids are globally unique and
// immutable), so it needs no server round-trip: it can be tested directly.
func TestDeriveIdentity(t *testing.T) {
	derived, err := Engine{}.DeriveIdentity(context.Background(), bigQueryConfig("my-gcp-project", "EU"))
	require.NoError(t, err)
	assert.Equal(t, "bigquery", derived.Engine)
	assert.True(t, strings.HasPrefix(derived.Key, "bq-"), "key %q should be bq-prefixed", derived.Key)
	assert.Equal(t, "my-gcp-project", derived.Attributes["project_id"])
	assert.Equal(t, "EU", derived.Attributes["location"])

	// A different project fingerprints differently, so a name repointed at another
	// project is detected.
	other, err := Engine{}.DeriveIdentity(context.Background(), bigQueryConfig("other-project", ""))
	require.NoError(t, err)
	assert.NotEqual(t, derived.Key, other.Key)
	assert.NotContains(t, other.Attributes, "location", "location is omitted when unset")
}

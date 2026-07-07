package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"core/engine/config"
	"core/engine/driver/mysql"
	"core/engine/driver/postgres"
	"core/e2e/harness"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deriveContext bounds an identity derivation in these tests.
func deriveContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func postgresDeriveConfig(host, database string) config.Database {
	return config.Database{
		Name:    "target",
		Type:    config.DatabaseTypePostgres,
		Details: &config.Details{Host: host, Port: 5432, Database: database, Username: "postgres", Password: "pw"},
		TLS:     &config.TLSSettings{Mode: "disable"},
	}
}

func mysqlDeriveConfig(host, database string) config.Database {
	return config.Database{
		Name:    "target",
		Type:    config.DatabaseTypeMySQL,
		Details: &config.Details{Host: host, Port: 3306, Database: database, Username: "root", Password: "pw"},
		TLS:     &config.TLSSettings{Mode: "disable"},
	}
}

// TestDeriveIdentityPostgres: the postgres driver fingerprints a real server, the
// fingerprint is stable across reconnects, and two distinct servers differ.
func TestDeriveIdentityPostgres(t *testing.T) {
	harness.RequireDocker(t)
	ctx := deriveContext(t)
	host := startPostgres(t)

	first, err := postgres.Engine{}.DeriveIdentity(ctx, postgresDeriveConfig(host, "cafe"))
	require.NoError(t, err)
	assert.Equal(t, "postgres", first.Engine)
	assert.True(t, strings.HasPrefix(first.Key, "pg-"), "key %q should be pg-prefixed", first.Key)
	assert.NotEmpty(t, first.Attributes["system_identifier"])
	assert.NotEmpty(t, first.Attributes["database_oid"])
	assert.Equal(t, "cafe", first.Attributes["database_name"])

	// Stable across a fresh connection to the same server.
	second, err := postgres.Engine{}.DeriveIdentity(ctx, postgresDeriveConfig(host, "cafe"))
	require.NoError(t, err)
	assert.Equal(t, first.Key, second.Key)

	// A separate server (its own initdb) has a different system_identifier.
	otherHost := startPostgres(t)
	other, err := postgres.Engine{}.DeriveIdentity(ctx, postgresDeriveConfig(otherHost, "cafe"))
	require.NoError(t, err)
	assert.NotEqual(t, first.Key, other.Key, "distinct servers must fingerprint differently")
}

// TestDeriveIdentityMySQL: the mysql driver fingerprints a real server via
// @@server_uuid + the active schema.
func TestDeriveIdentityMySQL(t *testing.T) {
	harness.RequireDocker(t)
	ctx := deriveContext(t)
	host := startMySQL(t)

	derived, err := mysql.NewMySQL().DeriveIdentity(ctx, mysqlDeriveConfig(host, "dbctx_test_dira"))
	require.NoError(t, err)
	assert.Equal(t, "mysql", derived.Engine)
	assert.True(t, strings.HasPrefix(derived.Key, "mysql-"), "key %q should be mysql-prefixed", derived.Key)
	assert.NotEmpty(t, derived.Attributes["server_uuid"])
	assert.Equal(t, "dbctx_test_dira", derived.Attributes["schema"])
}

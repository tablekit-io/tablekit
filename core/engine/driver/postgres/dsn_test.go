package postgres

import (
	"testing"

	"core/engine/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnConfigFromDetails(t *testing.T) {
	db := config.Database{
		Type:    config.DatabaseTypePostgres,
		Details: &config.Details{Host: "pg.internal", Port: 5432, Database: "app", Username: "app_ro", Password: "s3cret"},
	}
	cfg, serverHost, err := connConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "pg.internal", serverHost)
	assert.Equal(t, "pg.internal", cfg.Host)
	assert.Equal(t, uint16(5432), cfg.Port)
	assert.Equal(t, "app", cfg.Database)
	assert.Equal(t, "app_ro", cfg.User)
	assert.Equal(t, "s3cret", cfg.Password)
}

func TestConnConfigFromConnectionString(t *testing.T) {
	db := config.Database{
		Type:             config.DatabaseTypePostgres,
		ConnectionString: "postgres://reader@warehouse.internal:5433/reporting",
	}
	cfg, serverHost, err := connConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "warehouse.internal", serverHost)
	assert.Equal(t, "warehouse.internal", cfg.Host)
	assert.Equal(t, uint16(5433), cfg.Port)
	assert.Equal(t, "reporting", cfg.Database)
}

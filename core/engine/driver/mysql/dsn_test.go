package mysql

import (
	"strings"
	"testing"
	"time"

	"core/engine/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnConfigFromDetails(t *testing.T) {
	db := config.Database{
		Type:    config.DatabaseTypeMySQL,
		Details: &config.Details{Host: "my.internal", Port: 3306, Database: "shop", Username: "reader", Password: "pw"},
	}
	cfg, serverHost, err := connConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "my.internal", serverHost)
	assert.Equal(t, "my.internal:3306", cfg.Addr)
	assert.Equal(t, "tcp", cfg.Net)
	assert.Equal(t, "reader", cfg.User)
	assert.Equal(t, "pw", cfg.Passwd)
	assert.Equal(t, "shop", cfg.DBName)
}

func TestConnConfigFromConnectionString(t *testing.T) {
	db := config.Database{
		Type:             config.DatabaseTypeMySQL,
		ConnectionString: "mysql://reader:pw@my.internal:3307/shop?parseTime=true",
	}
	cfg, serverHost, err := connConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "my.internal", serverHost)
	assert.Equal(t, "my.internal:3307", cfg.Addr)
	assert.Equal(t, "reader", cfg.User)
	assert.Equal(t, "pw", cfg.Passwd)
	assert.Equal(t, "shop", cfg.DBName)
	assert.Equal(t, "true", cfg.Params["parseTime"])
}

func TestConnConfigConnectionStringDefaultPort(t *testing.T) {
	db := config.Database{
		Type:             config.DatabaseTypeMySQL,
		ConnectionString: "mysql://reader@my.internal/shop",
	}
	cfg, _, err := connConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "my.internal:3306", cfg.Addr)
}

func TestTimeoutStatements(t *testing.T) {
	timeout := 10 * time.Second

	mysqlStmt := NewMySQL().timeoutStatement(timeout)
	assert.Equal(t, "SET SESSION MAX_EXECUTION_TIME=10000", mysqlStmt)
	assert.True(t, strings.Contains(mysqlStmt, "MAX_EXECUTION_TIME"))

	mariaStmt := NewMariaDB().timeoutStatement(timeout)
	assert.Equal(t, "SET SESSION max_statement_time=10", mariaStmt)
	assert.True(t, strings.Contains(mariaStmt, "max_statement_time"))
}

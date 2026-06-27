package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMysqlConfigFromDetails(t *testing.T) {
	db := database{
		dbType:  databaseTypeMySQL,
		details: &details{host: "my.internal", port: 3306, database: "shop", username: "reader", password: "pw"},
	}
	config, serverHost, err := mysqlConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "my.internal", serverHost)
	assert.Equal(t, "my.internal:3306", config.Addr)
	assert.Equal(t, "tcp", config.Net)
	assert.Equal(t, "reader", config.User)
	assert.Equal(t, "pw", config.Passwd)
	assert.Equal(t, "shop", config.DBName)
}

func TestMysqlConfigFromConnectionString(t *testing.T) {
	db := database{
		dbType:           databaseTypeMySQL,
		connectionString: "mysql://reader:pw@my.internal:3307/shop?parseTime=true",
	}
	config, serverHost, err := mysqlConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "my.internal", serverHost)
	assert.Equal(t, "my.internal:3307", config.Addr)
	assert.Equal(t, "reader", config.User)
	assert.Equal(t, "pw", config.Passwd)
	assert.Equal(t, "shop", config.DBName)
	assert.Equal(t, "true", config.Params["parseTime"])
}

func TestMysqlConfigConnectionStringDefaultPort(t *testing.T) {
	db := database{
		dbType:           databaseTypeMySQL,
		connectionString: "mysql://reader@my.internal/shop",
	}
	config, _, err := mysqlConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "my.internal:3306", config.Addr)
}

func TestPgConnConfigFromDetails(t *testing.T) {
	db := database{
		dbType:  databaseTypePostgres,
		details: &details{host: "pg.internal", port: 5432, database: "app", username: "app_ro", password: "s3cret"},
	}
	config, serverHost, err := pgConnConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "pg.internal", serverHost)
	assert.Equal(t, "pg.internal", config.Host)
	assert.Equal(t, uint16(5432), config.Port)
	assert.Equal(t, "app", config.Database)
	assert.Equal(t, "app_ro", config.User)
	assert.Equal(t, "s3cret", config.Password)
}

func TestPgConnConfigFromConnectionString(t *testing.T) {
	db := database{
		dbType:           databaseTypePostgres,
		connectionString: "postgres://reader@warehouse.internal:5433/reporting",
	}
	config, serverHost, err := pgConnConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "warehouse.internal", serverHost)
	assert.Equal(t, "warehouse.internal", config.Host)
	assert.Equal(t, uint16(5433), config.Port)
	assert.Equal(t, "reporting", config.Database)
}

func TestTimeoutStatements(t *testing.T) {
	timeout := 10 * time.Second

	mysqlStmt := newMySQLEngine().timeoutStatement(timeout)
	assert.Equal(t, "SET SESSION MAX_EXECUTION_TIME=10000", mysqlStmt)
	assert.True(t, strings.Contains(mysqlStmt, "MAX_EXECUTION_TIME"))

	mariaStmt := newMariaDBEngine().timeoutStatement(timeout)
	assert.Equal(t, "SET SESSION max_statement_time=10", mariaStmt)
	assert.True(t, strings.Contains(mariaStmt, "max_statement_time"))
}

package engine

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// mysqlEngine runs read-only queries against MySQL and MariaDB over
// database/sql. The two flavours are wire-compatible and share this whole
// implementation; they differ only in the statement that arms the server-side
// query timeout, which is held as a function field (not a method override —
// Go embedding has no virtual dispatch, so an override would be silently
// skipped).
type mysqlEngine struct {
	timeoutStatement func(timeout time.Duration) string
}

func newMySQLEngine() mysqlEngine {
	return mysqlEngine{timeoutStatement: func(timeout time.Duration) string {
		return "SET SESSION MAX_EXECUTION_TIME=" + strconv.FormatInt(timeout.Milliseconds(), 10)
	}}
}

func newMariaDBEngine() mysqlEngine {
	return mysqlEngine{timeoutStatement: func(timeout time.Duration) string {
		return "SET SESSION max_statement_time=" + strconv.FormatFloat(timeout.Seconds(), 'f', -1, 64)
	}}
}

func (e mysqlEngine) run(ctx context.Context, db database, query string, limits Limits) (*Result, error) {
	config, serverHost, err := mysqlConfig(db)
	if err != nil {
		return nil, err
	}

	if db.ssh != nil {
		localAddr, cleanup, tunnelErr := openTunnel(db.ssh, config.Addr)
		if tunnelErr != nil {
			return nil, tunnelErr
		}
		defer cleanup()
		config.Addr = localAddr
		config.Net = "tcp"
	}

	tlsConfig, err := buildTLS(db.tls, serverHost)
	if err != nil {
		return nil, err
	}
	config.TLS = tlsConfig
	config.Timeout = sshDialTimeout

	ctx, cancel := context.WithTimeout(ctx, limits.StatementTimeout+connectGraceTimeout)
	defer cancel()

	connector, err := mysql.NewConnector(config)
	if err != nil {
		return nil, fmt.Errorf("build connector: %w", err)
	}
	pool := sql.OpenDB(connector)
	defer pool.Close()

	// Pin a single connection so the session timeout and the read-only
	// transaction run on the same physical connection.
	conn, err := pool.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	// Best-effort: not every server build supports the timeout variable.
	_, _ = conn.ExecContext(ctx, e.timeoutStatement(limits.StatementTimeout))

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin read-only transaction: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	accumulator := newRowAccumulator(columns, limits)
	scanTargets := make([]any, len(columns))
	scanPointers := make([]any, len(columns))
	for i := range scanTargets {
		scanPointers[i] = &scanTargets[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanPointers...); err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		row := make(map[string]any, len(columns))
		for i, column := range columns {
			normalized, keep, reason := normalizeValue(scanTargets[i])
			if !keep {
				accumulator.omit(column, reason)
				continue
			}
			row[column] = normalized
		}
		keepReading, addErr := accumulator.add(row)
		if addErr != nil {
			return nil, addErr
		}
		if !keepReading {
			break
		}
	}
	if err := rows.Err(); err != nil && !accumulator.truncated {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return accumulator.result(), nil
}

// mysqlConfig builds a go-sql-driver config from either structured details or a
// mysql:// connection URI, and returns the real server host for TLS/tunnel
// targeting.
func mysqlConfig(db database) (*mysql.Config, string, error) {
	config := mysql.NewConfig()
	config.Net = "tcp"

	if db.details != nil {
		config.User = db.details.username
		config.Passwd = db.details.password
		config.Addr = net.JoinHostPort(db.details.host, strconv.Itoa(db.details.port))
		config.DBName = db.details.database
		return config, db.details.host, nil
	}

	parsed, err := url.Parse(db.connectionString)
	if err != nil {
		return nil, "", fmt.Errorf("parse connection string: %w", err)
	}
	if parsed.User != nil {
		config.User = parsed.User.Username()
		if password, ok := parsed.User.Password(); ok {
			config.Passwd = password
		}
	}
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		port = "3306"
	}
	config.Addr = net.JoinHostPort(host, port)
	config.DBName = strings.TrimPrefix(parsed.Path, "/")
	if params := parsed.Query(); len(params) > 0 {
		config.Params = make(map[string]string, len(params))
		for key := range params {
			config.Params[key] = params.Get(key)
		}
	}
	return config, host, nil
}

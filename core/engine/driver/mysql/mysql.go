// Package mysql runs read-only queries against MySQL and MariaDB over
// database/sql. The two flavours are wire-compatible and share this whole
// implementation.
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"core/engine/config"
	"core/engine/encoding"
	"core/engine/transport/dbtls"
	"core/engine/transport/sshtunnel"

	gomysql "github.com/go-sql-driver/mysql"
)

// connectGraceTimeout is added on top of the statement timeout so the Go context
// deadline outlives the server-side cap; the server's timeout then fires first
// and yields a clean error rather than a context cancellation.
const connectGraceTimeout = 5 * time.Second

// Engine runs read-only queries against MySQL and MariaDB. The two flavours
// differ only in the statement that arms the server-side query timeout, which is
// held as a function field (not a method override — Go embedding has no virtual
// dispatch, so an override would be silently skipped).
type Engine struct {
	timeoutStatement func(timeout time.Duration) string
}

func NewMySQL() Engine {
	return Engine{timeoutStatement: func(timeout time.Duration) string {
		return "SET SESSION MAX_EXECUTION_TIME=" + strconv.FormatInt(timeout.Milliseconds(), 10)
	}}
}

func NewMariaDB() Engine {
	return Engine{timeoutStatement: func(timeout time.Duration) string {
		return "SET SESSION max_statement_time=" + strconv.FormatFloat(timeout.Seconds(), 'f', -1, 64)
	}}
}

func (e Engine) Run(ctx context.Context, db config.Database, query string, limits config.Limits) (*encoding.Result, error) {
	cfg, serverHost, err := connConfig(db)
	if err != nil {
		return nil, err
	}

	if db.SSH != nil {
		localAddr, cleanup, tunnelErr := sshtunnel.Open(db.SSH, cfg.Addr)
		if tunnelErr != nil {
			return nil, tunnelErr
		}
		defer cleanup()
		cfg.Addr = localAddr
		cfg.Net = "tcp"
	}

	tlsConfig, err := dbtls.BuildConfig(db.TLS, serverHost)
	if err != nil {
		return nil, err
	}
	cfg.TLS = tlsConfig
	cfg.Timeout = sshtunnel.DialTimeout

	ctx, cancel := context.WithTimeout(ctx, limits.StatementTimeout+connectGraceTimeout)
	defer cancel()

	connector, err := gomysql.NewConnector(cfg)
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

	accumulator := encoding.NewAccumulator(columns, limits.MaxRows, limits.MaxBytes)
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
			normalized, keep, reason := encoding.NormalizeValue(scanTargets[i])
			if !keep {
				accumulator.Omit(column, reason)
				continue
			}
			row[column] = normalized
		}
		keepReading, addErr := accumulator.Add(row)
		if addErr != nil {
			return nil, addErr
		}
		if !keepReading {
			break
		}
	}
	if err := rows.Err(); err != nil && !accumulator.Truncated() {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return accumulator.Result(), nil
}

// connConfig builds a go-sql-driver config from either structured details or a
// mysql:// connection URI, and returns the real server host for TLS/tunnel
// targeting.
func connConfig(db config.Database) (*gomysql.Config, string, error) {
	cfg := gomysql.NewConfig()
	cfg.Net = "tcp"

	if db.Details != nil {
		cfg.User = db.Details.Username
		cfg.Passwd = db.Details.Password
		cfg.Addr = net.JoinHostPort(db.Details.Host, strconv.Itoa(db.Details.Port))
		cfg.DBName = db.Details.Database
		return cfg, db.Details.Host, nil
	}

	parsed, err := url.Parse(db.ConnectionString)
	if err != nil {
		return nil, "", fmt.Errorf("parse connection string: %w", err)
	}
	if parsed.User != nil {
		cfg.User = parsed.User.Username()
		if password, ok := parsed.User.Password(); ok {
			cfg.Passwd = password
		}
	}
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		port = "3306"
	}
	cfg.Addr = net.JoinHostPort(host, port)
	cfg.DBName = strings.TrimPrefix(parsed.Path, "/")
	if params := parsed.Query(); len(params) > 0 {
		cfg.Params = make(map[string]string, len(params))
		for key := range params {
			cfg.Params[key] = params.Get(key)
		}
	}
	return cfg, host, nil
}

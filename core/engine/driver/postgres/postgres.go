// Package postgres runs read-only queries against PostgreSQL using pgx natively.
package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"core/engine/config"
	"core/engine/encoding"
	"core/engine/transport/dbtls"
	"core/engine/transport/sshtunnel"

	"github.com/jackc/pgx/v5"
)

// connectGraceTimeout is added on top of the statement timeout so the Go context
// deadline outlives the server-side cap; the server's statement_timeout then
// fires first and yields a clean error rather than a context cancellation.
const connectGraceTimeout = 5 * time.Second

// Engine runs read-only queries against PostgreSQL.
type Engine struct{}

func (Engine) Run(ctx context.Context, db config.Database, query string, limits config.Limits) (*encoding.Result, error) {
	connConfig, serverHost, err := connConfig(db)
	if err != nil {
		return nil, err
	}

	if db.SSH != nil {
		target := net.JoinHostPort(connConfig.Host, strconv.Itoa(int(connConfig.Port)))
		localAddr, cleanup, tunnelErr := sshtunnel.Open(db.SSH, target)
		if tunnelErr != nil {
			return nil, tunnelErr
		}
		defer cleanup()

		host, portStr, splitErr := net.SplitHostPort(localAddr)
		if splitErr != nil {
			return nil, splitErr
		}
		port, _ := strconv.Atoi(portStr)
		connConfig.Host = host
		connConfig.Port = uint16(port)
	}

	tlsConfig, err := dbtls.BuildConfig(db.TLS, serverHost)
	if err != nil {
		return nil, err
	}
	connConfig.TLSConfig = tlsConfig
	// Drop libpq-style fallbacks: each carries its own host/port/TLS and could
	// bypass the tunnel and our TLS settings by retrying the real address.
	connConfig.Fallbacks = nil

	ctx, cancel := context.WithTimeout(ctx, limits.StatementTimeout+connectGraceTimeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin read-only transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	timeoutMS := strconv.FormatInt(limits.StatementTimeout.Milliseconds(), 10)
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = "+timeoutMS); err != nil {
		return nil, fmt.Errorf("set statement timeout: %w", err)
	}

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, field := range fields {
		columns[i] = field.Name
	}

	accumulator := encoding.NewAccumulator(columns, limits.MaxRows, limits.MaxBytes)
	for rows.Next() {
		values, valuesErr := rows.Values()
		if valuesErr != nil {
			return nil, fmt.Errorf("read row: %w", valuesErr)
		}
		row := make(map[string]any, len(columns))
		for i, column := range columns {
			normalized, keep, reason := encoding.NormalizeValue(values[i])
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

// connConfig builds a pgx config from either structured details or a connection
// string, and returns the real server host for TLS/tunnel targeting. pgx
// requires configs to come from ParseConfig, so details are rendered into a URL
// first (which also keeps environment variables from leaking in).
func connConfig(db config.Database) (*pgx.ConnConfig, string, error) {
	if db.Details != nil {
		dsn := url.URL{
			Scheme: "postgres",
			Host:   net.JoinHostPort(db.Details.Host, strconv.Itoa(db.Details.Port)),
			Path:   "/" + db.Details.Database,
		}
		if db.Details.Password != "" {
			dsn.User = url.UserPassword(db.Details.Username, db.Details.Password)
		} else {
			dsn.User = url.User(db.Details.Username)
		}
		config, err := pgx.ParseConfig(dsn.String())
		if err != nil {
			return nil, "", fmt.Errorf("build connection config: %w", err)
		}
		return config, db.Details.Host, nil
	}

	config, err := pgx.ParseConfig(db.ConnectionString)
	if err != nil {
		return nil, "", fmt.Errorf("parse connection string: %w", err)
	}
	return config, config.Host, nil
}

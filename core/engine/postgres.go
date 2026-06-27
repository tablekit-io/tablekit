package engine

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// connectGraceTimeout is added on top of the statement timeout so the Go context
// deadline outlives the server-side cap; the server's statement_timeout then
// fires first and yields a clean error rather than a context cancellation.
const connectGraceTimeout = 5 * time.Second

// postgresEngine runs read-only queries against PostgreSQL using pgx natively.
type postgresEngine struct{}

func (postgresEngine) run(ctx context.Context, db database, query string, limits Limits) (*Result, error) {
	connConfig, serverHost, err := pgConnConfig(db)
	if err != nil {
		return nil, err
	}

	if db.ssh != nil {
		target := net.JoinHostPort(connConfig.Host, strconv.Itoa(int(connConfig.Port)))
		localAddr, cleanup, tunnelErr := openTunnel(db.ssh, target)
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

	tlsConfig, err := buildTLS(db.tls, serverHost)
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

	accumulator := newRowAccumulator(columns, limits)
	for rows.Next() {
		values, valuesErr := rows.Values()
		if valuesErr != nil {
			return nil, fmt.Errorf("read row: %w", valuesErr)
		}
		row := make(map[string]any, len(columns))
		for i, column := range columns {
			normalized, keep, reason := normalizeValue(values[i])
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

// pgConnConfig builds a pgx config from either structured details or a
// connection string, and returns the real server host for TLS/tunnel targeting.
// pgx requires configs to come from ParseConfig, so details are rendered into a
// URL first (which also keeps environment variables from leaking in).
func pgConnConfig(db database) (*pgx.ConnConfig, string, error) {
	if db.details != nil {
		dsn := url.URL{
			Scheme: "postgres",
			Host:   net.JoinHostPort(db.details.host, strconv.Itoa(db.details.port)),
			Path:   "/" + db.details.database,
		}
		if db.details.password != "" {
			dsn.User = url.UserPassword(db.details.username, db.details.password)
		} else {
			dsn.User = url.User(db.details.username)
		}
		config, err := pgx.ParseConfig(dsn.String())
		if err != nil {
			return nil, "", fmt.Errorf("build connection config: %w", err)
		}
		return config, db.details.host, nil
	}

	config, err := pgx.ParseConfig(db.connectionString)
	if err != nil {
		return nil, "", fmt.Errorf("parse connection string: %w", err)
	}
	return config, config.Host, nil
}

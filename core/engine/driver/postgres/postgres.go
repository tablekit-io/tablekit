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
	"core/engine/identity"
	"core/engine/transport/dbtls"
	"core/engine/transport/sshtunnel"

	"github.com/jackc/pgx/v5"
)

// connectGraceTimeout is added on top of the statement timeout so the Go context
// deadline outlives the server-side cap; the server's statement_timeout then
// fires first and yields a clean error rather than a context cancellation.
const connectGraceTimeout = 5 * time.Second

// deriveTimeout bounds identity derivation, which is a single trivial query and
// does not flow through the query Limits.
const deriveTimeout = 10 * time.Second

// Engine runs read-only queries against PostgreSQL.
type Engine struct{}

// dial builds the connection (structured details or DSN), opens the SSH tunnel
// and applies TLS, then connects — the exact prologue Run and DeriveIdentity
// share. The returned cleanup closes the connection and tears down the tunnel;
// callers must apply their own deadline to ctx first (it bounds the connect).
func dial(ctx context.Context, db config.Database) (*pgx.Conn, func(), error) {
	connConfig, serverHost, err := connConfig(db)
	if err != nil {
		return nil, nil, err
	}

	cleanups := []func(){}
	runCleanups := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	if db.SSH != nil {
		target := net.JoinHostPort(connConfig.Host, strconv.Itoa(int(connConfig.Port)))
		localAddr, tunnelCleanup, tunnelErr := sshtunnel.Open(db.SSH, target)
		if tunnelErr != nil {
			return nil, nil, tunnelErr
		}
		cleanups = append(cleanups, tunnelCleanup)

		host, portStr, splitErr := net.SplitHostPort(localAddr)
		if splitErr != nil {
			runCleanups()
			return nil, nil, splitErr
		}
		port, _ := strconv.Atoi(portStr)
		connConfig.Host = host
		connConfig.Port = uint16(port)
	}

	tlsConfig, err := dbtls.BuildConfig(db.TLS, serverHost)
	if err != nil {
		runCleanups()
		return nil, nil, err
	}
	connConfig.TLSConfig = tlsConfig
	// Drop libpq-style fallbacks: each carries its own host/port/TLS and could
	// bypass the tunnel and our TLS settings by retrying the real address.
	connConfig.Fallbacks = nil

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		runCleanups()
		return nil, nil, fmt.Errorf("connect: %w", err)
	}
	cleanups = append(cleanups, func() { conn.Close(context.Background()) })
	return conn, runCleanups, nil
}

func (Engine) Run(ctx context.Context, db config.Database, query string, page config.Page, limits config.Limits) (*encoding.Result, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, limits.StatementTimeout+connectGraceTimeout)
	defer cancel()

	conn, cleanup, err := dial(ctx, db)
	if err != nil {
		return nil, false, err
	}
	defer cleanup()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, false, fmt.Errorf("begin read-only transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	timeoutMS := strconv.FormatInt(limits.StatementTimeout.Milliseconds(), 10)
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = "+timeoutMS); err != nil {
		return nil, false, fmt.Errorf("set statement timeout: %w", err)
	}

	rows, err := tx.Query(ctx, pageQuery(query, page))
	if err != nil {
		return nil, false, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, field := range fields {
		columns[i] = field.Name
	}

	accumulator := encoding.NewAccumulator(columns, accumulatorRows(page), limits.MaxBytes)
	for rows.Next() {
		values, valuesErr := rows.Values()
		if valuesErr != nil {
			return nil, false, fmt.Errorf("read row: %w", valuesErr)
		}
		row := make(map[string]any, len(columns))
		for i, column := range columns {
			normalized, keep, reason := normalizeValue(values[i])
			if !keep {
				accumulator.Omit(column, reason)
				continue
			}
			row[column] = normalized
		}
		keepReading, addErr := accumulator.Add(row)
		if addErr != nil {
			return nil, false, addErr
		}
		if !keepReading {
			break
		}
	}
	if err := rows.Err(); err != nil && !accumulator.Truncated() {
		return nil, false, fmt.Errorf("iterate rows: %w", err)
	}

	result := accumulator.Result()
	hasMore := trimToLimit(result, page.Limit)
	return result, hasMore, nil
}

// DeriveIdentity fingerprints the physical PostgreSQL database behind db by
// reading the cluster's system_identifier plus the current database's oid and
// name. system_identifier is minted at initdb and survives restarts, so together
// with the database oid it pins the physical database independently of the
// connection details in databases.yaml.
func (Engine) DeriveIdentity(ctx context.Context, db config.Database) (identity.Identity, error) {
	ctx, cancel := context.WithTimeout(ctx, deriveTimeout+connectGraceTimeout)
	defer cancel()

	conn, cleanup, err := dial(ctx, db)
	if err != nil {
		return identity.Identity{}, err
	}
	defer cleanup()

	var (
		systemIdentifier string
		databaseOID      int
		databaseName     string
	)
	row := conn.QueryRow(ctx, `
		SELECT
			system_identifier::text,
			(SELECT oid::int FROM pg_database WHERE datname = current_database()) AS database_oid,
			current_database() AS database_name
		FROM pg_control_system()`)
	if err := row.Scan(&systemIdentifier, &databaseOID, &databaseName); err != nil {
		return identity.Identity{}, fmt.Errorf("derive identity: %w", err)
	}

	return identity.Identity{
		Engine: string(config.DatabaseTypePostgres),
		Key: fmt.Sprintf("pg-%s-%d",
			identity.Sanitize(systemIdentifier), databaseOID),
		Attributes: map[string]string{
			"system_identifier": systemIdentifier,
			"database_oid":      strconv.Itoa(databaseOID),
			"database_name":     databaseName,
		},
	}, nil
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

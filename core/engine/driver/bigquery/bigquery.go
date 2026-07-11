// Package bigquery runs read-only queries against Google BigQuery using the
// official cloud.google.com/go/bigquery client.
//
// BigQuery has no read-only transaction mode (its transactions only make
// multi-statement DML atomic), so read-only is enforced differently from the SQL
// drivers: every query is dry-run first and its StatementType inspected, and only
// a plain SELECT is allowed to run for real. The timeout is a job property rather
// than a session statement, and there is no cluster fingerprint — identity is the
// globally unique, immutable project id.
package bigquery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"core/engine/config"
	"core/engine/encoding"
	"core/engine/identity"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	bigqueryapi "cloud.google.com/go/bigquery"
)

// connectGraceTimeout is added on top of the statement timeout so the Go context
// deadline outlives the job's own timeout; the job-level timeout then fires first
// and yields a clean error rather than a context cancellation.
const connectGraceTimeout = 5 * time.Second

// deriveTimeout bounds identity derivation. Identity is derived from config
// without a round-trip, so this only guards the (currently trivial) work.
const deriveTimeout = 10 * time.Second

// selectStatementType is the only StatementType a query may have to run. A plain
// SELECT and a WITH ... SELECT both report as "SELECT"; anything else (INSERT,
// UPDATE, DELETE, CREATE_TABLE, SCRIPT, ...) is a write or multi-statement and is
// refused.
const selectStatementType = "SELECT"

// emulatorHostEnv is the standard BigQuery emulator variable. When set, the driver
// talks to the emulator at that address with authentication disabled instead of
// the real API — the seam the e2e suite uses. It is honored only in that env, so a
// production deployment that never sets it always takes the credentials-file path.
const emulatorHostEnv = "BIGQUERY_EMULATOR_HOST"

// newClient opens a BigQuery client for db. It is a package var so a test in this
// package could stub it. The returned cleanup closes the client; callers apply
// their own deadline to ctx first (it bounds client creation).
var newClient = func(ctx context.Context, db config.Database) (*bigqueryapi.Client, func(), error) {
	var options []option.ClientOption
	if endpoint := os.Getenv(emulatorHostEnv); endpoint != "" {
		options = append(options, option.WithEndpoint(endpoint), option.WithoutAuthentication())
	} else {
		options = append(options, option.WithCredentialsFile(db.BigQuery.CredentialsFilePath))
	}

	client, err := bigqueryapi.NewClient(ctx, db.BigQuery.ProjectID, options...)
	if err != nil {
		return nil, nil, fmt.Errorf("open bigquery client: %w", err)
	}
	return client, func() { _ = client.Close() }, nil
}

// Engine runs read-only queries against BigQuery.
type Engine struct{}

// Run applies the page window, verifies the query is read-only via a dry run, then
// executes it and accumulates the rows under the caller's caps.
func (Engine) Run(ctx context.Context, db config.Database, query string, page config.Page, limits config.Limits) (*encoding.Result, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, limits.StatementTimeout+connectGraceTimeout)
	defer cancel()

	client, cleanup, err := newClient(ctx, db)
	if err != nil {
		return nil, false, err
	}
	defer cleanup()

	paged := pageQuery(query, page)

	if err := ensureReadOnly(ctx, client, db, paged); err != nil {
		return nil, false, err
	}

	run := client.Query(paged)
	run.Location = db.BigQuery.Location
	run.JobTimeout = limits.StatementTimeout

	rows, err := run.Read(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("query: %w", err)
	}

	columns := columnNames(rows.Schema)
	accumulator := encoding.NewAccumulator(columns, accumulatorRows(page), limits.MaxBytes)
	for {
		var values []bigqueryapi.Value
		nextErr := rows.Next(&values)
		if errors.Is(nextErr, iterator.Done) {
			break
		}
		if nextErr != nil {
			return nil, false, fmt.Errorf("read row: %w", nextErr)
		}
		// The schema is only populated after the first Next; recompute columns once
		// it is available in case the first page was empty at Read time.
		if len(columns) == 0 {
			columns = columnNames(rows.Schema)
		}

		row := make(map[string]any, len(columns))
		for index, column := range columns {
			if index >= len(values) {
				break
			}
			normalized, keep, reason := normalizeValue(values[index], fieldAt(rows.Schema, index))
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

	result := accumulator.Result()
	result.Columns = columns
	hasMore := trimToLimit(result, page.Limit)
	return result, hasMore, nil
}

// ensureReadOnly dry-runs the query and refuses it unless its statement type is a
// plain SELECT. This is BigQuery's equivalent of the SQL drivers' read-only
// transaction: a dry run resolves the statement type without executing or billing.
func ensureReadOnly(ctx context.Context, client *bigqueryapi.Client, db config.Database, query string) error {
	dryRun := client.Query(query)
	dryRun.Location = db.BigQuery.Location
	dryRun.DryRun = true

	job, err := dryRun.Run(ctx)
	if err != nil {
		return fmt.Errorf("validate query: %w", err)
	}

	status := job.LastStatus()
	if status == nil || status.Statistics == nil {
		return errors.New("validate query: no statistics returned by dry run")
	}
	stats, ok := status.Statistics.Details.(*bigqueryapi.QueryStatistics)
	if !ok {
		return errors.New("validate query: dry run did not return query statistics")
	}
	return checkStatementType(stats.StatementType)
}

// checkStatementType is the read-only decision, split out from the dry-run
// plumbing so it can be tested without a BigQuery client. It permits only a plain
// SELECT (which a WITH ... SELECT also reports as); every write or multi-statement
// form is refused.
func checkStatementType(statementType string) error {
	if statementType != selectStatementType {
		return fmt.Errorf("only read-only SELECT queries are allowed; got statement type %q", statementType)
	}
	return nil
}

// DeriveIdentity fingerprints the BigQuery project behind db. Project ids are
// globally unique and immutable, so — unlike the SQL engines, which read a
// server/cluster identifier — the identity is derived from config without a
// round-trip. Re-pointing a name at a different project changes the id, which is
// exactly what the resolver needs to detect.
func (Engine) DeriveIdentity(_ context.Context, db config.Database) (identity.Identity, error) {
	attributes := map[string]string{"project_id": db.BigQuery.ProjectID}
	if db.BigQuery.Location != "" {
		attributes["location"] = db.BigQuery.Location
	}
	return identity.Identity{
		Engine:     string(config.DatabaseTypeBigQuery),
		Key:        "bq-" + identity.Sanitize(db.BigQuery.ProjectID),
		Attributes: attributes,
	}, nil
}

// columnNames extracts the column names from a result schema, in order.
func columnNames(schema bigqueryapi.Schema) []string {
	names := make([]string, len(schema))
	for index, field := range schema {
		names[index] = field.Name
	}
	return names
}

// fieldAt returns the schema field at index, or nil when the schema is shorter
// than the row (which should not happen, but keeps normalization safe).
func fieldAt(schema bigqueryapi.Schema, index int) *bigqueryapi.FieldSchema {
	if index < 0 || index >= len(schema) {
		return nil
	}
	return schema[index]
}

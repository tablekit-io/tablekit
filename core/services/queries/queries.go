// Package queries is the repository for query_database descriptors: the {database,
// sql, description} records that query_database saves and that read_results,
// fetch_chart_data, the chart tools and get_export_url later load by key to
// re-run against the live database. It stores descriptors only — never result
// rows — so the table stays small and every read reflects current data.
package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"core/db/gen/tablekit/public/model"
	"core/db/gen/tablekit/public/table"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"

	. "github.com/go-jet/jet/v2/postgres"
)

// Descriptor is one stored query: the key clients pass around, the database it
// targets, the read-only SQL to re-run, and the agent's plain-language intent.
type Descriptor struct {
	ID          string    `json:"id"`
	Database    string    `json:"database"`
	SQL         string    `json:"sql"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// QueryRepository persists and loads Descriptors in the mcp_queries table.
type QueryRepository interface {
	Save(ctx context.Context, database, query, description string) (string, error)
	Get(ctx context.Context, id string) (*Descriptor, error)
}

type queryRepository struct {
	database *sql.DB
}

// New returns a QueryRepository over the given database. The schema is owned by
// the db package's migrations; this type only reads and writes rows.
func New(database *sql.DB) QueryRepository {
	return &queryRepository{database: database}
}

// Save inserts a new descriptor and returns its generated key. The key is a
// UUIDv7 (time-ordered), so keys sort by creation and don't leak a sequence.
func (r *queryRepository) Save(ctx context.Context, database, query, description string) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate query id: %w", err)
	}
	key := id.String()
	stmt := table.McpQueries.
		INSERT(table.McpQueries.ID, table.McpQueries.Database, table.McpQueries.SQL, table.McpQueries.Description).
		VALUES(key, database, query, description)
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return "", fmt.Errorf("save query: %w", err)
	}
	return key, nil
}

// Get loads a descriptor by key. A key that doesn't exist is not an error: it
// returns (nil, nil) so callers can turn an unknown key into a tool-level
// message rather than a server error.
func (r *queryRepository) Get(ctx context.Context, id string) (*Descriptor, error) {
	stmt := SELECT(table.McpQueries.AllColumns).
		FROM(table.McpQueries).
		WHERE(table.McpQueries.ID.EQ(String(id)))

	var row model.McpQueries
	err := stmt.QueryContext(ctx, r.database, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get query %q: %w", id, err)
	}
	return &Descriptor{
		ID:          row.ID,
		Database:    row.Database,
		SQL:         row.SQL,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
	}, nil
}

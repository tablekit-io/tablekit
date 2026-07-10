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

// Descriptor is one stored query: the key clients pass around, the physical
// database it targets (by database_id, not the mutable databases.yaml name), the
// read-only SQL to re-run, and the agent's plain-language intent.
type Descriptor struct {
	ID          uuid.UUID `json:"id"`
	DatabaseID  uuid.UUID `json:"database_id"`
	SQL         string    `json:"sql"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// QueryRepository persists and loads Descriptors in the queries table.
type QueryRepository interface {
	Save(ctx context.Context, databaseID, clientID uuid.UUID, query, description string) (uuid.UUID, error)
	Get(ctx context.Context, id uuid.UUID) (*Descriptor, error)
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
// clientID is the client that created the query, so every stored query is
// attributable.
func (r *queryRepository) Save(ctx context.Context, databaseID, clientID uuid.UUID, query, description string) (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate query id: %w", err)
	}
	stmt := table.Queries.
		INSERT(table.Queries.ID, table.Queries.DatabaseID, table.Queries.ClientID, table.Queries.SQL, table.Queries.Description).
		VALUES(UUID(id), UUID(databaseID), UUID(clientID), query, description)
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return uuid.Nil, fmt.Errorf("save query: %w", err)
	}
	return id, nil
}

// Get loads a descriptor by key. A key that doesn't exist is not an error: it
// returns (nil, nil) so callers can turn an unknown key into a tool-level
// message rather than a server error.
func (r *queryRepository) Get(ctx context.Context, id uuid.UUID) (*Descriptor, error) {
	stmt := SELECT(table.Queries.AllColumns).
		FROM(table.Queries).
		WHERE(table.Queries.ID.EQ(UUID(id)))

	var row model.Queries
	err := stmt.QueryContext(ctx, r.database, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get query %q: %w", id, err)
	}
	return &Descriptor{
		ID:          row.ID,
		DatabaseID:  row.DatabaseID,
		SQL:         row.SQL,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
	}, nil
}

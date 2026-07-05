// Package queries is the repository for run_query descriptors: the {database,
// sql, description} records that run_query saves and that retrieve_results,
// fetch_chart_data, render_*_chart and get_export_url later load by key to
// re-run against the live database. It stores descriptors only — never result
// rows — so the table stays small and every read reflects current data.
package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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

// Repository persists and loads Descriptors in the mcp_queries table.
type Repository struct {
	database *sql.DB
}

// New returns a Repository over the given database. The schema is owned by the
// db package's migrations; this type only reads and writes rows.
func New(database *sql.DB) *Repository {
	return &Repository{database: database}
}

// Save inserts a new descriptor and returns its generated key. The key is a
// UUIDv7 (time-ordered), so keys sort by creation and don't leak a sequence.
func (r *Repository) Save(ctx context.Context, database, query, description string) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate query id: %w", err)
	}
	key := id.String()
	_, err = r.database.ExecContext(ctx,
		`INSERT INTO mcp_queries (id, database, sql, description) VALUES ($1, $2, $3, $4)`,
		key, database, query, description,
	)
	if err != nil {
		return "", fmt.Errorf("save query: %w", err)
	}
	return key, nil
}

// Get loads a descriptor by key. A key that doesn't exist is not an error: it
// returns (nil, nil) so callers can turn an unknown key into a tool-level
// message rather than a server error.
func (r *Repository) Get(ctx context.Context, id string) (*Descriptor, error) {
	row := r.database.QueryRowContext(ctx,
		`SELECT id, database, sql, description, created_at FROM mcp_queries WHERE id = $1`,
		id,
	)
	var descriptor Descriptor
	err := row.Scan(
		&descriptor.ID,
		&descriptor.Database,
		&descriptor.SQL,
		&descriptor.Description,
		&descriptor.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get query %q: %w", id, err)
	}
	return &descriptor, nil
}

// Package databases is the repository and resolver for physical-database
// identity. The databases table holds one row per physical database tablekit has
// queried, keyed by a driver-derived fingerprint (identity_key) rather than the
// databases.yaml name. The Resolver turns a configured name into a stable
// database_id on first query (deriving and persisting the identity), caches it,
// and — on re-run — refuses when a name now points at a different physical
// database than the stored query was saved against.
package databases

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"core/db/gen/tablekit/public/model"
	"core/db/gen/tablekit/public/table"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"

	. "github.com/go-jet/jet/v2/postgres"
)

// Record is one databases row: the stable database_id, the name last seen under,
// the engine family, the static fingerprint used for matching, and the
// structured fingerprint kept for observability.
type Record struct {
	ID          uuid.UUID
	Name        string
	Type        string
	IdentityKey string
	Identity    map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Repository persists and loads databases rows.
type Repository interface {
	// Upsert inserts a row for the given id, or — when identity_key already
	// exists — updates the stored name/identity and returns the STABLE existing
	// id. So a physical database keeps one database_id even as the name it is
	// configured under changes, and concurrent first queries converge on one row.
	Upsert(ctx context.Context, id uuid.UUID, record Record) (uuid.UUID, error)
	// Get loads a row by database_id, returning (nil, nil) when absent.
	Get(ctx context.Context, id uuid.UUID) (*Record, error)
}

type repository struct {
	database *sql.DB
}

// NewRepository returns a Repository over the given database. The schema is owned
// by the db package's migrations; this type only reads and writes rows.
func NewRepository(database *sql.DB) Repository {
	return &repository{database: database}
}

func (r *repository) Upsert(ctx context.Context, id uuid.UUID, record Record) (uuid.UUID, error) {
	identityJSON, err := json.Marshal(record.Identity)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encode identity: %w", err)
	}

	// identity is sent as text and cast into the jsonb column by Postgres, the
	// same pattern the config store uses. On an identity_key conflict the name
	// and identity are refreshed and updated_at is bumped to the proposed row's
	// default (CURRENT_TIMESTAMP), while RETURNING yields the pre-existing id.
	stmt := table.Databases.
		INSERT(
			table.Databases.ID, table.Databases.Name, table.Databases.Type,
			table.Databases.IdentityKey, table.Databases.Identity,
		).
		VALUES(UUID(id), record.Name, record.Type, record.IdentityKey, string(identityJSON)).
		ON_CONFLICT(table.Databases.IdentityKey).
		DO_UPDATE(SET(
			table.Databases.Name.SET(table.Databases.EXCLUDED.Name),
			table.Databases.Identity.SET(table.Databases.EXCLUDED.Identity),
			table.Databases.UpdatedAt.SET(table.Databases.EXCLUDED.UpdatedAt),
		)).
		RETURNING(table.Databases.ID)

	var row model.Databases
	if err := stmt.QueryContext(ctx, r.database, &row); err != nil {
		return uuid.Nil, fmt.Errorf("upsert database %q: %w", record.IdentityKey, err)
	}
	return row.ID, nil
}

func (r *repository) Get(ctx context.Context, id uuid.UUID) (*Record, error) {
	stmt := SELECT(table.Databases.AllColumns).
		FROM(table.Databases).
		WHERE(table.Databases.ID.EQ(UUID(id)))

	var row model.Databases
	err := stmt.QueryContext(ctx, r.database, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get database %q: %w", id, err)
	}
	return &Record{
		ID:          row.ID,
		Name:        row.Name,
		Type:        string(row.Type),
		IdentityKey: row.IdentityKey,
		Identity:    row.Identity.Val,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

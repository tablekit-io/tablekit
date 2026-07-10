package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"core/db/dbjson"
	"core/db/gen/tablekit/public/model"
	"core/db/gen/tablekit/public/table"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"

	. "github.com/go-jet/jet/v2/postgres"
)

// Client types, matching the client_type enum. An OAuth client is a dynamic
// registration; a static client owns a CLI-minted static token.
const (
	ClientTypeOAuth  = "oauth"
	ClientTypeStatic = "static"
)

// Client is a registered client in the unified clients table. It is either an
// OAuth client (RFC 7591 dynamic registration, Type "oauth") or a static-token
// client (a CLI-minted static token registers its own client, Type "static" with
// a nil ClientName and empty RedirectURIs).
type Client struct {
	ClientID uuid.UUID `json:"client_id"`
	// ClientName is a pointer so an absent name is stored as NULL rather than an
	// empty string, which is what static clients carry.
	ClientName   *string  `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
	// Type distinguishes a static-token client ("static") from an OAuth client
	// ("oauth"). Maps to the client_type enum column.
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// ClientRepository persists registered clients (both OAuth and static-token
// clients) in the clients table.
type ClientRepository interface {
	SaveClient(ctx context.Context, c *Client) error
	GetClient(ctx context.Context, id uuid.UUID) (*Client, error)
}

type clientRepository struct {
	database *sql.DB
}

// NewClientRepository returns a ClientRepository over the given database.
func NewClientRepository(database *sql.DB) ClientRepository {
	return &clientRepository{database: database}
}

// SaveClient persists a newly registered client. redirect_uris is stored as a
// jsonb array (read/written whole via the typed dbjson.JSON wrapper).
func (r *clientRepository) SaveClient(ctx context.Context, c *Client) error {
	stmt := table.Clients.
		INSERT(table.Clients.AllColumns).
		MODEL(model.Clients{
			ID:           c.ClientID,
			ClientName:   c.ClientName,
			Type:         model.ClientType(c.Type),
			RedirectUris: dbjson.JSON[[]string]{Val: c.RedirectURIs},
			CreatedAt:    c.CreatedAt,
		})
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("save client: %w", err)
	}
	return nil
}

// GetClient returns the client by id, or nil if unknown.
func (r *clientRepository) GetClient(ctx context.Context, id uuid.UUID) (*Client, error) {
	stmt := SELECT(table.Clients.AllColumns).
		FROM(table.Clients).
		WHERE(table.Clients.ID.EQ(UUID(id)))

	var row model.Clients
	err := stmt.QueryContext(ctx, r.database, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get client %q: %w", id, err)
	}
	return &Client{
		ClientID:     row.ID,
		ClientName:   row.ClientName,
		RedirectURIs: row.RedirectUris.Val,
		Type:         string(row.Type),
		CreatedAt:    row.CreatedAt,
	}, nil
}

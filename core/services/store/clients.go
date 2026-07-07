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

// Client is a registered OAuth client (RFC 7591 dynamic registration). A CLI-
// minted bearer token also registers a Client, with Type "bearer", a nil
// ClientName (stored as SQL NULL) and an empty RedirectURIs.
type Client struct {
	ClientID uuid.UUID `json:"client_id"`
	// ClientName is a pointer so an absent name is stored as NULL rather than an
	// empty string, which is what bearer clients carry.
	ClientName   *string  `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
	// Type distinguishes a CLI bearer client ("bearer") from an OAuth client
	// (empty/omitted).
	Type      string    `json:"type,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ClientRepository persists registered OAuth clients (and CLI bearer clients) in
// the oauth_clients table.
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
	stmt := table.OAuthClients.
		INSERT(table.OAuthClients.AllColumns).
		MODEL(model.OAuthClients{
			ClientID:     c.ClientID,
			ClientName:   c.ClientName,
			Type:         c.Type,
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
	stmt := SELECT(table.OAuthClients.AllColumns).
		FROM(table.OAuthClients).
		WHERE(table.OAuthClients.ClientID.EQ(UUID(id)))

	var row model.OAuthClients
	err := stmt.QueryContext(ctx, r.database, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get client %q: %w", id, err)
	}
	return &Client{
		ClientID:     row.ClientID,
		ClientName:   row.ClientName,
		RedirectURIs: row.RedirectUris.Val,
		Type:         row.Type,
		CreatedAt:    row.CreatedAt,
	}, nil
}

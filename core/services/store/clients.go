package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Client is a registered OAuth client (RFC 7591 dynamic registration). A CLI-
// minted bearer token also registers a Client, with Type "bearer", a nil
// ClientName (stored as SQL NULL) and an empty RedirectURIs.
type Client struct {
	ClientID string `json:"client_id"`
	// ClientName is a pointer so an absent name is stored as NULL rather than an
	// empty string, which is what bearer clients carry.
	ClientName   *string  `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
	// Type distinguishes a CLI bearer client ("bearer") from an OAuth client
	// (empty/omitted).
	Type      string    `json:"type,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Pairing modes control whether a not-yet-paired client may pair.
const (
	PairingOnce       = "once"       // next new client pairs, then mode flips to disabled
	PairingIndefinite = "indefinite" // every new client may pair
	PairingDisabled   = "disabled"   // no new client may pair
)

// pairingModeKey is the config row that holds the current pairing mode.
const pairingModeKey = "pairing_mode"

// SaveClient persists a newly registered client. redirect_uris is stored as a
// JSON array in a single column (read/written whole).
func (s *Store) SaveClient(ctx context.Context, c *Client) error {
	redirectURIs, err := json.Marshal(c.RedirectURIs)
	if err != nil {
		return fmt.Errorf("marshal redirect_uris: %w", err)
	}
	_, err = s.database.ExecContext(ctx,
		`INSERT INTO oauth_clients (client_id, client_name, type, redirect_uris, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		c.ClientID, c.ClientName, c.Type, string(redirectURIs), c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save client: %w", err)
	}
	return nil
}

// GetClient returns the client by id, or nil if unknown.
func (s *Store) GetClient(ctx context.Context, id string) (*Client, error) {
	row := s.database.QueryRowContext(ctx,
		`SELECT client_id, client_name, type, redirect_uris, created_at
		 FROM oauth_clients WHERE client_id = $1`,
		id,
	)
	var (
		client       Client
		redirectURIs string
	)
	err := row.Scan(&client.ClientID, &client.ClientName, &client.Type, &redirectURIs, &client.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get client %q: %w", id, err)
	}
	if err := json.Unmarshal([]byte(redirectURIs), &client.RedirectURIs); err != nil {
		return nil, fmt.Errorf("unmarshal redirect_uris for %q: %w", id, err)
	}
	return &client, nil
}

// TryPair reports whether clientID may use the server, pairing it if the current
// mode allows. Already-paired clients are always allowed. The read-then-write is
// done in a transaction so concurrent authorizes cannot race the mode flip.
//
//   - disabled:   new clients rejected
//   - once:       new client paired, then mode flips to disabled
//   - indefinite: every new client paired; mode unchanged
func (s *Store) TryPair(ctx context.Context, clientID string) (bool, error) {
	// Serialize the read-check-write in-process so concurrent pairings under
	// "once" cannot both win a snapshot race (single-instance server).
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var existing string
	err = tx.QueryRowContext(ctx,
		`SELECT client_id FROM oauth_paired_clients WHERE client_id = $1`, clientID,
	).Scan(&existing)
	if err == nil {
		return true, nil // already paired
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("check paired client: %w", err)
	}

	mode, err := pairingMode(ctx, tx)
	if err != nil {
		return false, err
	}
	switch mode {
	case PairingIndefinite:
		if err := pairClient(ctx, tx, clientID); err != nil {
			return false, err
		}
	case PairingOnce:
		if err := pairClient(ctx, tx, clientID); err != nil {
			return false, err
		}
		if err := setPairingMode(ctx, tx, PairingDisabled); err != nil {
			return false, err
		}
	default: // disabled or unknown
		return false, nil
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

// SetPairingMode persists a new pairing mode. Used by the `pairing` CLI.
func (s *Store) SetPairingMode(ctx context.Context, mode string) error {
	switch mode {
	case PairingOnce, PairingIndefinite, PairingDisabled:
	default:
		return fmt.Errorf("unknown pairing mode %q", mode)
	}
	return setPairingMode(ctx, s.database, mode)
}

// PairingStatus returns the current mode and the paired client ids.
func (s *Store) PairingStatus(ctx context.Context) (mode string, paired []string, err error) {
	mode, err = pairingMode(ctx, s.database)
	if err != nil {
		return "", nil, err
	}
	rows, err := s.database.QueryContext(ctx,
		`SELECT client_id FROM oauth_paired_clients ORDER BY client_id`)
	if err != nil {
		return "", nil, fmt.Errorf("list paired clients: %w", err)
	}
	defer rows.Close()
	paired = []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", nil, err
		}
		paired = append(paired, id)
	}
	if err := rows.Err(); err != nil {
		return "", nil, err
	}
	return mode, paired, nil
}

// ---- helpers (work against either *sql.DB or *sql.Tx) -------------------

// querier is the subset of *sql.DB / *sql.Tx these helpers need.
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// getConfig reads the JSON value under key into dest, reporting whether the row
// existed. A missing key is not an error: it returns (false, nil) so callers can
// apply their own default.
func getConfig(ctx context.Context, q querier, key string, dest any) (bool, error) {
	var raw string
	err := q.QueryRowContext(ctx, `SELECT value FROM config WHERE key = $1`, key).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read config %q: %w", key, err)
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return false, fmt.Errorf("decode config %q: %w", key, err)
	}
	return true, nil
}

// setConfig upserts value (JSON-encoded) under key.
func setConfig(ctx context.Context, q querier, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode config %q: %w", key, err)
	}
	_, err = q.ExecContext(ctx,
		`INSERT INTO config (key, value) VALUES ($1, $2)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, string(raw),
	)
	if err != nil {
		return fmt.Errorf("set config %q: %w", key, err)
	}
	return nil
}

// pairingMode reads the current mode, defaulting to PairingOnce when the config
// row is absent (matching the old JSON default).
func pairingMode(ctx context.Context, q querier) (string, error) {
	var mode string
	found, err := getConfig(ctx, q, pairingModeKey, &mode)
	if err != nil {
		return "", err
	}
	if !found {
		return PairingOnce, nil
	}
	return mode, nil
}

// setPairingMode upserts the pairing_mode config value.
func setPairingMode(ctx context.Context, q querier, mode string) error {
	return setConfig(ctx, q, pairingModeKey, mode)
}

// pairClient adds clientID to the paired set (idempotent).
func pairClient(ctx context.Context, q querier, clientID string) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO oauth_paired_clients (client_id) VALUES ($1)
		 ON CONFLICT(client_id) DO NOTHING`,
		clientID,
	)
	if err != nil {
		return fmt.Errorf("pair client: %w", err)
	}
	return nil
}

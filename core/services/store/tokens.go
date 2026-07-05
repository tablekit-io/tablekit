package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// AuthCode is a one-time authorization code bound to a PKCE challenge.
type AuthCode struct {
	Code          string    `json:"code"`
	ClientID      string    `json:"client_id"`
	RedirectURI   string    `json:"redirect_uri"`
	CodeChallenge string    `json:"code_challenge"`
	Scope         string    `json:"scope"`
	UserID        string    `json:"user_id"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// Chain tracks a refresh-token lineage. Rotation bumps InvalidatedBefore to the
// iat of the just-used refresh token, so any older (already-rotated) refresh
// token is rejected as a replay — and a replay revokes the whole chain.
type Chain struct {
	ID                string    `json:"id"`
	ClientID          string    `json:"client_id"`
	UserID            string    `json:"user_id"`
	Scope             string    `json:"scope"`
	RedirectURI       string    `json:"redirect_uri"`
	Revoked           bool      `json:"revoked"`
	InvalidatedBefore time.Time `json:"invalidated_before"`
	CreatedAt         time.Time `json:"created_at"`
}

// BearerToken is a long-lived, CLI-minted access token. Unlike OAuth access
// tokens (which are short-lived and never persisted), a bearer token is recorded
// here so it can be revoked: the MCP guard looks it up by its jti on every
// request and rejects it once Revoked is set.
type BearerToken struct {
	ID        string    `json:"id"` // jti; links the JWT to this row
	ClientID  string    `json:"client_id"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ---- auth codes ---------------------------------------------------------

// PutCode stores a one-time authorization code.
func (s *Store) PutCode(ctx context.Context, c *AuthCode) error {
	_, err := s.database.ExecContext(ctx,
		`INSERT INTO oauth_auth_codes
		 (code, client_id, redirect_uri, code_challenge, scope, user_id, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		c.Code, c.ClientID, c.RedirectURI, c.CodeChallenge, c.Scope, c.UserID, c.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("put auth code: %w", err)
	}
	return nil
}

// ConsumeCode atomically fetches and deletes a code (single use). Returns nil if
// the code is unknown/already used. The SELECT + DELETE run in a transaction so
// two redemptions of the same code cannot both succeed.
func (s *Store) ConsumeCode(ctx context.Context, code string) (*AuthCode, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var c AuthCode
	err = tx.QueryRowContext(ctx,
		`SELECT code, client_id, redirect_uri, code_challenge, scope, user_id, expires_at
		 FROM oauth_auth_codes WHERE code = $1`, code,
	).Scan(&c.Code, &c.ClientID, &c.RedirectURI, &c.CodeChallenge, &c.Scope, &c.UserID, &c.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read auth code: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM oauth_auth_codes WHERE code = $1`, code); err != nil {
		return nil, fmt.Errorf("consume auth code: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &c, nil
}

// ---- chains -------------------------------------------------------------

// NewChain persists a fresh refresh-token chain.
func (s *Store) NewChain(ctx context.Context, c *Chain) error {
	_, err := s.database.ExecContext(ctx,
		`INSERT INTO oauth_token_chains
		 (id, client_id, user_id, scope, redirect_uri, revoked, invalidated_before, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.ClientID, c.UserID, c.Scope, c.RedirectURI, c.Revoked, c.InvalidatedBefore, c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("new chain: %w", err)
	}
	return nil
}

// GetChain returns the chain by id, or nil if unknown.
func (s *Store) GetChain(ctx context.Context, id string) (*Chain, error) {
	row := s.database.QueryRowContext(ctx,
		`SELECT id, client_id, user_id, scope, redirect_uri, revoked, invalidated_before, created_at
		 FROM oauth_token_chains WHERE id = $1`, id,
	)
	var c Chain
	err := row.Scan(&c.ID, &c.ClientID, &c.UserID, &c.Scope, &c.RedirectURI,
		&c.Revoked, &c.InvalidatedBefore, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get chain %q: %w", id, err)
	}
	return &c, nil
}

// BumpCutoff advances a chain's InvalidatedBefore to t (rotation).
func (s *Store) BumpCutoff(ctx context.Context, id string, t time.Time) error {
	_, err := s.database.ExecContext(ctx,
		`UPDATE oauth_token_chains SET invalidated_before = $1 WHERE id = $2`, t, id,
	)
	if err != nil {
		return fmt.Errorf("bump chain cutoff %q: %w", id, err)
	}
	return nil
}

// RevokeChain marks a chain revoked (replay detected / logout).
func (s *Store) RevokeChain(ctx context.Context, id string) error {
	_, err := s.database.ExecContext(ctx,
		`UPDATE oauth_token_chains SET revoked = TRUE WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("revoke chain %q: %w", id, err)
	}
	return nil
}

// ---- bearer tokens ------------------------------------------------------

// PutBearerToken persists a CLI-minted bearer token.
func (s *Store) PutBearerToken(ctx context.Context, t *BearerToken) error {
	_, err := s.database.ExecContext(ctx,
		`INSERT INTO oauth_bearer_tokens (id, client_id, revoked, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		t.ID, t.ClientID, t.Revoked, t.CreatedAt, t.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("put bearer token: %w", err)
	}
	return nil
}

// GetBearerToken returns the bearer token by id, or nil if unknown.
func (s *Store) GetBearerToken(ctx context.Context, id string) (*BearerToken, error) {
	row := s.database.QueryRowContext(ctx,
		`SELECT id, client_id, revoked, created_at, expires_at
		 FROM oauth_bearer_tokens WHERE id = $1`, id,
	)
	var t BearerToken
	err := row.Scan(&t.ID, &t.ClientID, &t.Revoked, &t.CreatedAt, &t.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get bearer token %q: %w", id, err)
	}
	return &t, nil
}

// RevokeBearerToken marks a bearer token revoked. It returns an error if the id
// is unknown, so the CLI can tell the user nothing was revoked.
func (s *Store) RevokeBearerToken(ctx context.Context, id string) error {
	result, err := s.database.ExecContext(ctx,
		`UPDATE oauth_bearer_tokens SET revoked = TRUE WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("revoke bearer token %q: %w", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no bearer token with id %q", id)
	}
	return nil
}

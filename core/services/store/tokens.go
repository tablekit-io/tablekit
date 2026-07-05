package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"database/sql"

	"core/db/gen/tablekit/public/model"
	"core/db/gen/tablekit/public/table"

	"github.com/go-jet/jet/v2/qrm"

	. "github.com/go-jet/jet/v2/postgres"
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

// AuthCodeRepository persists one-time PKCE authorization codes.
type AuthCodeRepository interface {
	PutCode(ctx context.Context, c *AuthCode) error
	ConsumeCode(ctx context.Context, code string) (*AuthCode, error)
}

type authCodeRepository struct {
	database *sql.DB
}

// NewAuthCodeRepository returns an AuthCodeRepository over the given database.
func NewAuthCodeRepository(database *sql.DB) AuthCodeRepository {
	return &authCodeRepository{database: database}
}

// PutCode stores a one-time authorization code.
func (r *authCodeRepository) PutCode(ctx context.Context, c *AuthCode) error {
	stmt := table.OAuthAuthCodes.
		INSERT(table.OAuthAuthCodes.AllColumns).
		MODEL(model.OAuthAuthCodes{
			Code:          c.Code,
			ClientID:      c.ClientID,
			RedirectURI:   c.RedirectURI,
			CodeChallenge: c.CodeChallenge,
			Scope:         c.Scope,
			UserID:        c.UserID,
			ExpiresAt:     c.ExpiresAt,
		})
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("put auth code: %w", err)
	}
	return nil
}

// ConsumeCode atomically fetches and deletes a code (single use). Returns nil if
// the code is unknown/already used. The SELECT + DELETE run in a transaction so
// two redemptions of the same code cannot both succeed.
func (r *authCodeRepository) ConsumeCode(ctx context.Context, code string) (*AuthCode, error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var row model.OAuthAuthCodes
	err = SELECT(table.OAuthAuthCodes.AllColumns).
		FROM(table.OAuthAuthCodes).
		WHERE(table.OAuthAuthCodes.Code.EQ(String(code))).
		QueryContext(ctx, tx, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read auth code: %w", err)
	}

	if _, err := table.OAuthAuthCodes.
		DELETE().
		WHERE(table.OAuthAuthCodes.Code.EQ(String(code))).
		ExecContext(ctx, tx); err != nil {
		return nil, fmt.Errorf("consume auth code: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &AuthCode{
		Code:          row.Code,
		ClientID:      row.ClientID,
		RedirectURI:   row.RedirectURI,
		CodeChallenge: row.CodeChallenge,
		Scope:         row.Scope,
		UserID:        row.UserID,
		ExpiresAt:     row.ExpiresAt,
	}, nil
}

// ---- chains -------------------------------------------------------------

// TokenChainRepository persists refresh-token lineages.
type TokenChainRepository interface {
	NewChain(ctx context.Context, c *Chain) error
	GetChain(ctx context.Context, id string) (*Chain, error)
	BumpCutoff(ctx context.Context, id string, t time.Time) error
	RevokeChain(ctx context.Context, id string) error
}

type tokenChainRepository struct {
	database *sql.DB
}

// NewTokenChainRepository returns a TokenChainRepository over the given database.
func NewTokenChainRepository(database *sql.DB) TokenChainRepository {
	return &tokenChainRepository{database: database}
}

// NewChain persists a fresh refresh-token chain.
func (r *tokenChainRepository) NewChain(ctx context.Context, c *Chain) error {
	stmt := table.OAuthTokenChains.
		INSERT(table.OAuthTokenChains.AllColumns).
		MODEL(model.OAuthTokenChains{
			ID:                c.ID,
			ClientID:          c.ClientID,
			UserID:            c.UserID,
			Scope:             c.Scope,
			RedirectURI:       c.RedirectURI,
			Revoked:           c.Revoked,
			InvalidatedBefore: c.InvalidatedBefore,
			CreatedAt:         c.CreatedAt,
		})
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("new chain: %w", err)
	}
	return nil
}

// GetChain returns the chain by id, or nil if unknown.
func (r *tokenChainRepository) GetChain(ctx context.Context, id string) (*Chain, error) {
	stmt := SELECT(table.OAuthTokenChains.AllColumns).
		FROM(table.OAuthTokenChains).
		WHERE(table.OAuthTokenChains.ID.EQ(String(id)))

	var row model.OAuthTokenChains
	err := stmt.QueryContext(ctx, r.database, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get chain %q: %w", id, err)
	}
	return &Chain{
		ID:                row.ID,
		ClientID:          row.ClientID,
		UserID:            row.UserID,
		Scope:             row.Scope,
		RedirectURI:       row.RedirectURI,
		Revoked:           row.Revoked,
		InvalidatedBefore: row.InvalidatedBefore,
		CreatedAt:         row.CreatedAt,
	}, nil
}

// BumpCutoff advances a chain's InvalidatedBefore to t (rotation).
func (r *tokenChainRepository) BumpCutoff(ctx context.Context, id string, t time.Time) error {
	stmt := table.OAuthTokenChains.
		UPDATE(table.OAuthTokenChains.InvalidatedBefore).
		SET(TimestampzT(t)).
		WHERE(table.OAuthTokenChains.ID.EQ(String(id)))
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("bump chain cutoff %q: %w", id, err)
	}
	return nil
}

// RevokeChain marks a chain revoked (replay detected / logout).
func (r *tokenChainRepository) RevokeChain(ctx context.Context, id string) error {
	stmt := table.OAuthTokenChains.
		UPDATE(table.OAuthTokenChains.Revoked).
		SET(Bool(true)).
		WHERE(table.OAuthTokenChains.ID.EQ(String(id)))
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("revoke chain %q: %w", id, err)
	}
	return nil
}

// ---- bearer tokens ------------------------------------------------------

// BearerTokenRepository persists CLI-minted long-lived bearer tokens.
type BearerTokenRepository interface {
	PutBearerToken(ctx context.Context, t *BearerToken) error
	GetBearerToken(ctx context.Context, id string) (*BearerToken, error)
	RevokeBearerToken(ctx context.Context, id string) error
}

type bearerTokenRepository struct {
	database *sql.DB
}

// NewBearerTokenRepository returns a BearerTokenRepository over the given database.
func NewBearerTokenRepository(database *sql.DB) BearerTokenRepository {
	return &bearerTokenRepository{database: database}
}

// PutBearerToken persists a CLI-minted bearer token.
func (r *bearerTokenRepository) PutBearerToken(ctx context.Context, t *BearerToken) error {
	stmt := table.OAuthBearerTokens.
		INSERT(table.OAuthBearerTokens.AllColumns).
		MODEL(model.OAuthBearerTokens{
			ID:        t.ID,
			ClientID:  t.ClientID,
			Revoked:   t.Revoked,
			CreatedAt: t.CreatedAt,
			ExpiresAt: t.ExpiresAt,
		})
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("put bearer token: %w", err)
	}
	return nil
}

// GetBearerToken returns the bearer token by id, or nil if unknown.
func (r *bearerTokenRepository) GetBearerToken(ctx context.Context, id string) (*BearerToken, error) {
	stmt := SELECT(table.OAuthBearerTokens.AllColumns).
		FROM(table.OAuthBearerTokens).
		WHERE(table.OAuthBearerTokens.ID.EQ(String(id)))

	var row model.OAuthBearerTokens
	err := stmt.QueryContext(ctx, r.database, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get bearer token %q: %w", id, err)
	}
	return &BearerToken{
		ID:        row.ID,
		ClientID:  row.ClientID,
		Revoked:   row.Revoked,
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

// RevokeBearerToken marks a bearer token revoked. It returns an error if the id
// is unknown, so the CLI can tell the user nothing was revoked.
func (r *bearerTokenRepository) RevokeBearerToken(ctx context.Context, id string) error {
	result, err := table.OAuthBearerTokens.
		UPDATE(table.OAuthBearerTokens.Revoked).
		SET(Bool(true)).
		WHERE(table.OAuthBearerTokens.ID.EQ(String(id))).
		ExecContext(ctx, r.database)
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

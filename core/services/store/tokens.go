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
	"github.com/google/uuid"

	. "github.com/go-jet/jet/v2/postgres"
)

// AuthCode is a one-time authorization code bound to a PKCE challenge. ID is the
// row's primary key; Code is the opaque value handed to the client in the
// redirect and looked up on redemption. Today they carry the same value.
type AuthCode struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	ClientID      uuid.UUID `json:"client_id"`
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
	ID                uuid.UUID  `json:"id"`
	ClientID          uuid.UUID  `json:"client_id"`
	UserID            string     `json:"user_id"`
	Scope             string     `json:"scope"`
	RedirectURI       string     `json:"redirect_uri"`
	RevokedAt         *time.Time `json:"revoked_at"`
	InvalidatedBefore time.Time  `json:"invalidated_before"`
	CreatedAt         time.Time  `json:"created_at"`
}

// Revoked reports whether the chain has been revoked.
func (c *Chain) Revoked() bool { return c.RevokedAt != nil }

// StaticToken is a long-lived, CLI-minted token. Unlike OAuth access tokens
// (which are short-lived and never persisted), a static token is recorded here so
// it can be revoked: the MCP guard looks it up by its jti on every request and
// rejects it once RevokedAt is set.
type StaticToken struct {
	ID        uuid.UUID  `json:"id"` // jti; links the JWT to this row
	ClientID  uuid.UUID  `json:"client_id"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
}

// Revoked reports whether the token has been revoked.
func (t *StaticToken) Revoked() bool { return t.RevokedAt != nil }

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
			ID:            c.ID,
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

// ConsumeCode atomically fetches and deletes a code by its value (single use).
// Returns nil if the code is unknown/already used. The SELECT + DELETE run in a
// transaction so two redemptions of the same code cannot both succeed. The lookup
// is by the indexed code value; the delete targets the row's primary key.
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
		WHERE(table.OAuthAuthCodes.ID.EQ(UUID(row.ID))).
		ExecContext(ctx, tx); err != nil {
		return nil, fmt.Errorf("consume auth code: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &AuthCode{
		ID:            row.ID,
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
	GetChain(ctx context.Context, id uuid.UUID) (*Chain, error)
	BumpCutoff(ctx context.Context, id uuid.UUID, t time.Time) error
	RevokeChain(ctx context.Context, id uuid.UUID) error
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
			RevokedAt:         c.RevokedAt,
			InvalidatedBefore: c.InvalidatedBefore,
			CreatedAt:         c.CreatedAt,
		})
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("new chain: %w", err)
	}
	return nil
}

// GetChain returns the chain by id, or nil if unknown.
func (r *tokenChainRepository) GetChain(ctx context.Context, id uuid.UUID) (*Chain, error) {
	stmt := SELECT(table.OAuthTokenChains.AllColumns).
		FROM(table.OAuthTokenChains).
		WHERE(table.OAuthTokenChains.ID.EQ(UUID(id)))

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
		RevokedAt:         row.RevokedAt,
		InvalidatedBefore: row.InvalidatedBefore,
		CreatedAt:         row.CreatedAt,
	}, nil
}

// BumpCutoff advances a chain's InvalidatedBefore to t (rotation).
func (r *tokenChainRepository) BumpCutoff(ctx context.Context, id uuid.UUID, t time.Time) error {
	stmt := table.OAuthTokenChains.
		UPDATE(table.OAuthTokenChains.InvalidatedBefore).
		SET(TimestampzT(t)).
		WHERE(table.OAuthTokenChains.ID.EQ(UUID(id)))
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("bump chain cutoff %q: %w", id, err)
	}
	return nil
}

// RevokeChain marks a chain revoked (replay detected / logout) by stamping
// revoked_at with the current time.
func (r *tokenChainRepository) RevokeChain(ctx context.Context, id uuid.UUID) error {
	stmt := table.OAuthTokenChains.
		UPDATE(table.OAuthTokenChains.RevokedAt).
		SET(TimestampzT(time.Now())).
		WHERE(table.OAuthTokenChains.ID.EQ(UUID(id)))
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("revoke chain %q: %w", id, err)
	}
	return nil
}

// ---- static tokens ------------------------------------------------------

// StaticTokenRepository persists CLI-minted long-lived static tokens.
type StaticTokenRepository interface {
	PutStaticToken(ctx context.Context, t *StaticToken) error
	GetStaticToken(ctx context.Context, id uuid.UUID) (*StaticToken, error)
	RevokeStaticToken(ctx context.Context, id uuid.UUID) error
}

type staticTokenRepository struct {
	database *sql.DB
}

// NewStaticTokenRepository returns a StaticTokenRepository over the given database.
func NewStaticTokenRepository(database *sql.DB) StaticTokenRepository {
	return &staticTokenRepository{database: database}
}

// PutStaticToken persists a CLI-minted static token.
func (r *staticTokenRepository) PutStaticToken(ctx context.Context, t *StaticToken) error {
	stmt := table.StaticTokens.
		INSERT(table.StaticTokens.AllColumns).
		MODEL(model.StaticTokens{
			ID:        t.ID,
			ClientID:  t.ClientID,
			RevokedAt: t.RevokedAt,
			CreatedAt: t.CreatedAt,
			ExpiresAt: t.ExpiresAt,
		})
	if _, err := stmt.ExecContext(ctx, r.database); err != nil {
		return fmt.Errorf("put static token: %w", err)
	}
	return nil
}

// GetStaticToken returns the static token by id, or nil if unknown.
func (r *staticTokenRepository) GetStaticToken(ctx context.Context, id uuid.UUID) (*StaticToken, error) {
	stmt := SELECT(table.StaticTokens.AllColumns).
		FROM(table.StaticTokens).
		WHERE(table.StaticTokens.ID.EQ(UUID(id)))

	var row model.StaticTokens
	err := stmt.QueryContext(ctx, r.database, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get static token %q: %w", id, err)
	}
	return &StaticToken{
		ID:        row.ID,
		ClientID:  row.ClientID,
		RevokedAt: row.RevokedAt,
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

// RevokeStaticToken marks a static token revoked by stamping revoked_at. It
// returns an error if the id is unknown, so the CLI can tell the user nothing was
// revoked.
func (r *staticTokenRepository) RevokeStaticToken(ctx context.Context, id uuid.UUID) error {
	result, err := table.StaticTokens.
		UPDATE(table.StaticTokens.RevokedAt).
		SET(TimestampzT(time.Now())).
		WHERE(table.StaticTokens.ID.EQ(UUID(id))).
		ExecContext(ctx, r.database)
	if err != nil {
		return fmt.Errorf("revoke static token %q: %w", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no static token with id %q", id)
	}
	return nil
}

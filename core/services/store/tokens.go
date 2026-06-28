package store

import (
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

// tokensFile is the on-disk shape of tokens.json.
type tokensFile struct {
	Codes  map[string]*AuthCode    `json:"codes"`
	Chains map[string]*Chain       `json:"chains"`
	Tokens map[string]*BearerToken `json:"tokens"`
}

func (s *Store) loadTokens() (*tokensFile, error) {
	tokensData := &tokensFile{
		Codes:  map[string]*AuthCode{},
		Chains: map[string]*Chain{},
		Tokens: map[string]*BearerToken{},
	}
	if err := s.readJSON("tokens.json", tokensData); err != nil {
		return nil, err
	}
	if tokensData.Codes == nil {
		tokensData.Codes = map[string]*AuthCode{}
	}
	if tokensData.Chains == nil {
		tokensData.Chains = map[string]*Chain{}
	}
	if tokensData.Tokens == nil {
		tokensData.Tokens = map[string]*BearerToken{}
	}
	return tokensData, nil
}

// ---- auth codes ---------------------------------------------------------

// PutCode stores a one-time authorization code.
func (s *Store) PutCode(c *AuthCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokensData, err := s.loadTokens()
	if err != nil {
		return err
	}
	tokensData.Codes[c.Code] = c
	return s.writeJSON("tokens.json", tokensData)
}

// ConsumeCode atomically fetches and deletes a code (single use). Returns nil
// if the code is unknown/already used.
func (s *Store) ConsumeCode(code string) (*AuthCode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokensData, err := s.loadTokens()
	if err != nil {
		return nil, err
	}
	c := tokensData.Codes[code]
	if c == nil {
		return nil, nil
	}
	delete(tokensData.Codes, code)
	if err := s.writeJSON("tokens.json", tokensData); err != nil {
		return nil, err
	}
	return c, nil
}

// ---- chains -------------------------------------------------------------

// NewChain persists a fresh refresh-token chain.
func (s *Store) NewChain(c *Chain) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokensData, err := s.loadTokens()
	if err != nil {
		return err
	}
	tokensData.Chains[c.ID] = c
	return s.writeJSON("tokens.json", tokensData)
}

// GetChain returns the chain by id, or nil if unknown.
func (s *Store) GetChain(id string) (*Chain, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokensData, err := s.loadTokens()
	if err != nil {
		return nil, err
	}
	return tokensData.Chains[id], nil
}

// BumpCutoff advances a chain's InvalidatedBefore to t (rotation).
func (s *Store) BumpCutoff(id string, t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokensData, err := s.loadTokens()
	if err != nil {
		return err
	}
	if chain := tokensData.Chains[id]; chain != nil {
		chain.InvalidatedBefore = t
	}
	return s.writeJSON("tokens.json", tokensData)
}

// RevokeChain marks a chain revoked (replay detected / logout).
func (s *Store) RevokeChain(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokensData, err := s.loadTokens()
	if err != nil {
		return err
	}
	if chain := tokensData.Chains[id]; chain != nil {
		chain.Revoked = true
	}
	return s.writeJSON("tokens.json", tokensData)
}

// ---- bearer tokens ------------------------------------------------------

// PutBearerToken persists a CLI-minted bearer token.
func (s *Store) PutBearerToken(t *BearerToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokensData, err := s.loadTokens()
	if err != nil {
		return err
	}
	tokensData.Tokens[t.ID] = t
	return s.writeJSON("tokens.json", tokensData)
}

// GetBearerToken returns the bearer token by id, or nil if unknown.
func (s *Store) GetBearerToken(id string) (*BearerToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokensData, err := s.loadTokens()
	if err != nil {
		return nil, err
	}
	return tokensData.Tokens[id], nil
}

// RevokeBearerToken marks a bearer token revoked. It returns an error if the id
// is unknown, so the CLI can tell the user nothing was revoked.
func (s *Store) RevokeBearerToken(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokensData, err := s.loadTokens()
	if err != nil {
		return err
	}
	token := tokensData.Tokens[id]
	if token == nil {
		return fmt.Errorf("no bearer token with id %q", id)
	}
	token.Revoked = true
	return s.writeJSON("tokens.json", tokensData)
}

// Package store is the JSON-file persistence layer for the OAuth server.
//
// State is split across three gitignored files in DataDir:
//   - clients.json  registered clients + which one is "paired" (single-client lock)
//   - tokens.json   one-time auth codes + refresh-token chains (rotation/replay)
//   - signing.key   the HS256 secret (generated on first use)
//
// There is no database; this models the dbctx Postgres tables (oauth_clients,
// oauth_auth_codes, oauth_token_chains) as flat JSON. All mutations take a
// single process-wide mutex and persist with an atomic temp-file rename, which
// is sufficient for a single-instance, single-client server.
package store

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// Client is a registered OAuth client (RFC 7591 dynamic registration). A CLI-
// minted bearer token also registers a Client, with Type "bearer", a nil
// ClientName (serialized as null) and an empty RedirectURIs.
type Client struct {
	ClientID string `json:"client_id"`
	// ClientName is a pointer so an absent name serializes as JSON null rather
	// than being omitted, which is what bearer clients carry.
	ClientName   *string   `json:"client_name"`
	RedirectURIs []string  `json:"redirect_uris"`
	// Type distinguishes a CLI bearer client ("bearer") from an OAuth client
	// (empty/omitted).
	Type      string    `json:"type,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

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

// Pairing modes control whether a not-yet-paired client may pair.
const (
	PairingOnce       = "once"       // next new client pairs, then mode flips to disabled
	PairingIndefinite = "indefinite" // every new client may pair
	PairingDisabled   = "disabled"   // no new client may pair
)

// clientsFile is the on-disk shape of clients.json.
type clientsFile struct {
	// PairingMode gates whether new clients may pair; defaults to "once".
	PairingMode string `json:"pairing_mode"`
	// PairedClientIDs are the clients allowed to authenticate.
	PairedClientIDs []string `json:"paired_client_ids"`
	// LegacyPairedClientID migrates the old single-client field: it is folded
	// into PairedClientIDs on load and dropped on the next save (omitempty).
	LegacyPairedClientID string             `json:"paired_client_id,omitempty"`
	Clients              map[string]*Client `json:"clients"`
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

// Store is the persistence handle. Construct with New.
type Store struct {
	directory string
	mu        sync.Mutex
}

// New ensures DataDir exists and returns a Store. Existing state files are
// loaded once up front so schema violations fail fast at startup rather than on
// the first request that touches them.
func New(directory string) (*Store, error) {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	s := &Store{directory: directory}
	if _, err := s.loadClients(); err != nil {
		return nil, fmt.Errorf("loading clients.json: %w", err)
	}
	if _, err := s.loadTokens(); err != nil {
		return nil, fmt.Errorf("loading tokens.json: %w", err)
	}
	if err := s.migrateLegacySigningKey(); err != nil {
		return nil, fmt.Errorf("migrating signing.key: %w", err)
	}
	return s, nil
}

// ---- low-level file helpers (callers hold s.mu) -------------------------

func (s *Store) path(name string) string { return filepath.Join(s.directory, name) }

// readJSON loads name into v. A missing file is not an error: v keeps its
// zero/initialized value so callers can start from empty state.
func (s *Store) readJSON(name string, v any) error {
	b, err := os.ReadFile(s.path(name))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	if err := validateAgainstSchema(name, b); err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// writeJSON atomically persists v to name (write temp, fsync-free rename).
func (s *Store) writeJSON(name string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tempPath := s.path(name) + ".tmp"
	if err := os.WriteFile(tempPath, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tempPath, s.path(name))
}

func (s *Store) loadClients() (*clientsFile, error) {
	clientsData := &clientsFile{Clients: map[string]*Client{}}
	if err := s.readJSON("clients.json", clientsData); err != nil {
		return nil, err
	}
	if clientsData.Clients == nil {
		clientsData.Clients = map[string]*Client{}
	}
	if clientsData.PairedClientIDs == nil {
		// Keep it a non-nil slice so it marshals to [] not null (schema: array).
		clientsData.PairedClientIDs = []string{}
	}
	if clientsData.PairingMode == "" {
		clientsData.PairingMode = PairingOnce
	}
	// Fold the legacy single paired client into the list.
	if clientsData.LegacyPairedClientID != "" {
		if !slices.Contains(clientsData.PairedClientIDs, clientsData.LegacyPairedClientID) {
			clientsData.PairedClientIDs = append(clientsData.PairedClientIDs, clientsData.LegacyPairedClientID)
		}
		clientsData.LegacyPairedClientID = ""
	}
	return clientsData, nil
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

// ---- signing key --------------------------------------------------------

// signingKeyMinLen is the minimum HS256 key length; shorter keys are zero-padded
// up to it (256-bit keys for HMAC-SHA256).
const signingKeyMinLen = 32

// DecodeSigningKey decodes a base64 signing key (standard or raw, with or
// without `=` padding) and zero-pads it up to signingKeyMinLen. Keys at or above
// the minimum are returned as-is (never truncated). Empty or invalid input is an
// error. Shared by the env-provided key path and the on-disk key file.
func DecodeSigningKey(b64 string) ([]byte, error) {
	raw, err := decodeBase64Tolerant(strings.TrimSpace(b64))
	if err != nil {
		return nil, fmt.Errorf("invalid base64 signing key: %w", err)
	}
	if len(raw) == 0 {
		return nil, errors.New("signing key is empty")
	}
	if len(raw) >= signingKeyMinLen {
		return raw, nil
	}
	padded := make([]byte, signingKeyMinLen)
	copy(padded, raw)
	return padded, nil
}

// decodeBase64Tolerant accepts standard and URL alphabets, with or without
// padding.
func decodeBase64Tolerant(s string) ([]byte, error) {
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding,
		base64.URLEncoding, base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		}
	}
	return nil, errors.New("not valid base64")
}

// writeKeyFile persists raw key bytes as base64 text (0600).
func (s *Store) writeKeyFile(raw []byte) error {
	encoded := base64.StdEncoding.EncodeToString(raw)
	return os.WriteFile(s.path("signing.key"), []byte(encoded), 0o600)
}

// migrateLegacySigningKey runs at boot: a signing.key that is exactly 32 raw
// bytes is the pre-base64 format, so re-encode it in place. A missing file is
// left alone (an env key may supply the secret); anything else is assumed to be
// base64 already.
func (s *Store) migrateLegacySigningKey() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path("signing.key"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(b) == signingKeyMinLen {
		return s.writeKeyFile(b)
	}
	return nil
}

// SigningKey returns the HS256 secret from signing.key, generating and
// persisting a random 32-byte key (base64-encoded) on first call so tokens
// survive restarts with zero config. Legacy raw files are already normalized to
// base64 by migrateLegacySigningKey at boot, so this only sees base64 or absent.
func (s *Store) SigningKey() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path("signing.key"))
	if err == nil {
		return DecodeSigningKey(string(b))
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	key := make([]byte, signingKeyMinLen)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := s.writeKeyFile(key); err != nil {
		return nil, err
	}
	return key, nil
}

// ---- clients ------------------------------------------------------------

// SaveClient persists a newly registered client.
func (s *Store) SaveClient(c *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return err
	}
	clientsData.Clients[c.ClientID] = c
	return s.writeJSON("clients.json", clientsData)
}

// GetClient returns the client by id, or nil if unknown.
func (s *Store) GetClient(id string) (*Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return nil, err
	}
	return clientsData.Clients[id], nil
}

// TryPair reports whether clientID may use the server, pairing it if the
// current mode allows. Already-paired clients are always allowed.
//
//   - disabled:   new clients rejected
//   - once:       new client paired, then mode flips to disabled
//   - indefinite: every new client paired; mode unchanged
func (s *Store) TryPair(clientID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return false, err
	}
	if slices.Contains(clientsData.PairedClientIDs, clientID) {
		return true, nil
	}
	switch clientsData.PairingMode {
	case PairingIndefinite:
		clientsData.PairedClientIDs = append(clientsData.PairedClientIDs, clientID)
	case PairingOnce:
		clientsData.PairedClientIDs = append(clientsData.PairedClientIDs, clientID)
		clientsData.PairingMode = PairingDisabled
	default: // disabled or unknown
		return false, nil
	}
	if err := s.writeJSON("clients.json", clientsData); err != nil {
		return false, err
	}
	return true, nil
}

// SetPairingMode persists a new pairing mode. Used by the `pairing` CLI.
func (s *Store) SetPairingMode(mode string) error {
	switch mode {
	case PairingOnce, PairingIndefinite, PairingDisabled:
	default:
		return fmt.Errorf("unknown pairing mode %q", mode)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return err
	}
	clientsData.PairingMode = mode
	return s.writeJSON("clients.json", clientsData)
}

// PairingStatus returns the current mode and the paired client ids.
func (s *Store) PairingStatus() (mode string, paired []string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return "", nil, err
	}
	return clientsData.PairingMode, clientsData.PairedClientIDs, nil
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

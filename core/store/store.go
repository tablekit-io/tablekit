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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

// Client is a registered OAuth client (RFC 7591 dynamic registration).
type Client struct {
	ClientID     string    `json:"client_id"`
	ClientName   string    `json:"client_name,omitempty"`
	RedirectURIs []string  `json:"redirect_uris"`
	CreatedAt    time.Time `json:"created_at"`
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

// tokensFile is the on-disk shape of tokens.json.
type tokensFile struct {
	Codes  map[string]*AuthCode `json:"codes"`
	Chains map[string]*Chain    `json:"chains"`
}

// Store is the persistence handle. Construct with New.
type Store struct {
	dir string
	mu  sync.Mutex
}

// New ensures DataDir exists and returns a Store. Existing state files are
// loaded once up front so schema violations fail fast at startup rather than on
// the first request that touches them.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &Store{dir: dir}
	if _, err := s.loadClients(); err != nil {
		return nil, fmt.Errorf("loading clients.json: %w", err)
	}
	if _, err := s.loadTokens(); err != nil {
		return nil, fmt.Errorf("loading tokens.json: %w", err)
	}
	return s, nil
}

// ---- low-level file helpers (callers hold s.mu) -------------------------

func (s *Store) path(name string) string { return filepath.Join(s.dir, name) }

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
	tmp := s.path(name) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path(name))
}

func (s *Store) loadClients() (*clientsFile, error) {
	cf := &clientsFile{Clients: map[string]*Client{}}
	if err := s.readJSON("clients.json", cf); err != nil {
		return nil, err
	}
	if cf.Clients == nil {
		cf.Clients = map[string]*Client{}
	}
	if cf.PairedClientIDs == nil {
		// Keep it a non-nil slice so it marshals to [] not null (schema: array).
		cf.PairedClientIDs = []string{}
	}
	if cf.PairingMode == "" {
		cf.PairingMode = PairingOnce
	}
	// Fold the legacy single paired client into the list.
	if cf.LegacyPairedClientID != "" {
		if !slices.Contains(cf.PairedClientIDs, cf.LegacyPairedClientID) {
			cf.PairedClientIDs = append(cf.PairedClientIDs, cf.LegacyPairedClientID)
		}
		cf.LegacyPairedClientID = ""
	}
	return cf, nil
}

func (s *Store) loadTokens() (*tokensFile, error) {
	tf := &tokensFile{Codes: map[string]*AuthCode{}, Chains: map[string]*Chain{}}
	if err := s.readJSON("tokens.json", tf); err != nil {
		return nil, err
	}
	if tf.Codes == nil {
		tf.Codes = map[string]*AuthCode{}
	}
	if tf.Chains == nil {
		tf.Chains = map[string]*Chain{}
	}
	return tf, nil
}

// ---- signing key --------------------------------------------------------

// SigningKey returns the HS256 secret, generating and persisting a random
// 32-byte key on first call so tokens survive restarts with zero config.
func (s *Store) SigningKey() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path("signing.key"))
	if err == nil && len(b) >= 32 {
		return b, nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(s.path("signing.key"), key, 0o600); err != nil {
		return nil, err
	}
	return key, nil
}

// ---- clients ------------------------------------------------------------

// SaveClient persists a newly registered client.
func (s *Store) SaveClient(c *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cf, err := s.loadClients()
	if err != nil {
		return err
	}
	cf.Clients[c.ClientID] = c
	return s.writeJSON("clients.json", cf)
}

// GetClient returns the client by id, or nil if unknown.
func (s *Store) GetClient(id string) (*Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cf, err := s.loadClients()
	if err != nil {
		return nil, err
	}
	return cf.Clients[id], nil
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
	cf, err := s.loadClients()
	if err != nil {
		return false, err
	}
	if slices.Contains(cf.PairedClientIDs, clientID) {
		return true, nil
	}
	switch cf.PairingMode {
	case PairingIndefinite:
		cf.PairedClientIDs = append(cf.PairedClientIDs, clientID)
	case PairingOnce:
		cf.PairedClientIDs = append(cf.PairedClientIDs, clientID)
		cf.PairingMode = PairingDisabled
	default: // disabled or unknown
		return false, nil
	}
	if err := s.writeJSON("clients.json", cf); err != nil {
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
	cf, err := s.loadClients()
	if err != nil {
		return err
	}
	cf.PairingMode = mode
	return s.writeJSON("clients.json", cf)
}

// PairingStatus returns the current mode and the paired client ids.
func (s *Store) PairingStatus() (mode string, paired []string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cf, err := s.loadClients()
	if err != nil {
		return "", nil, err
	}
	return cf.PairingMode, cf.PairedClientIDs, nil
}

// ---- auth codes ---------------------------------------------------------

// PutCode stores a one-time authorization code.
func (s *Store) PutCode(c *AuthCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tf, err := s.loadTokens()
	if err != nil {
		return err
	}
	tf.Codes[c.Code] = c
	return s.writeJSON("tokens.json", tf)
}

// ConsumeCode atomically fetches and deletes a code (single use). Returns nil
// if the code is unknown/already used.
func (s *Store) ConsumeCode(code string) (*AuthCode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tf, err := s.loadTokens()
	if err != nil {
		return nil, err
	}
	c := tf.Codes[code]
	if c == nil {
		return nil, nil
	}
	delete(tf.Codes, code)
	if err := s.writeJSON("tokens.json", tf); err != nil {
		return nil, err
	}
	return c, nil
}

// ---- chains -------------------------------------------------------------

// NewChain persists a fresh refresh-token chain.
func (s *Store) NewChain(c *Chain) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tf, err := s.loadTokens()
	if err != nil {
		return err
	}
	tf.Chains[c.ID] = c
	return s.writeJSON("tokens.json", tf)
}

// GetChain returns the chain by id, or nil if unknown.
func (s *Store) GetChain(id string) (*Chain, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tf, err := s.loadTokens()
	if err != nil {
		return nil, err
	}
	return tf.Chains[id], nil
}

// BumpCutoff advances a chain's InvalidatedBefore to t (rotation).
func (s *Store) BumpCutoff(id string, t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tf, err := s.loadTokens()
	if err != nil {
		return err
	}
	if ch := tf.Chains[id]; ch != nil {
		ch.InvalidatedBefore = t
	}
	return s.writeJSON("tokens.json", tf)
}

// RevokeChain marks a chain revoked (replay detected / logout).
func (s *Store) RevokeChain(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tf, err := s.loadTokens()
	if err != nil {
		return err
	}
	if ch := tf.Chains[id]; ch != nil {
		ch.Revoked = true
	}
	return s.writeJSON("tokens.json", tf)
}

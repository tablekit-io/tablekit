// Package store is the JSON-file persistence layer for the OAuth server.
//
// State is split across three gitignored files in DataDir, each with its own
// source file here:
//   - clients.json  registered clients + pairing (clients.go)
//   - tokens.json   one-time auth codes, refresh chains, bearer tokens (tokens.go)
//   - signing.key   the HS256 secret, generated on first use (signing.go)
//
// There is no database; this models the dbctx Postgres tables (oauth_clients,
// oauth_auth_codes, oauth_token_chains) as flat JSON. All mutations take a
// single process-wide mutex and persist with an atomic temp-file rename, which
// is sufficient for a single-instance, single-client server.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

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

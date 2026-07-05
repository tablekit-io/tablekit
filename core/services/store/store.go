// Package store is the persistence layer for the OAuth/MCP server. It reads and
// writes the oauth_* tables in tablekit's SQLite database (schema owned by the
// db package's goose migrations) and, separately, the HS256 signing key.
//
// State that lives in SQLite:
//   - oauth_clients        registered clients + CLI bearer clients (clients.go)
//   - oauth_paired_clients + oauth_settings  pairing set + mode  (clients.go)
//   - oauth_auth_codes     one-time PKCE auth codes               (tokens.go)
//   - oauth_token_chains   refresh-token lineages                 (tokens.go)
//   - oauth_bearer_tokens  long-lived CLI bearer tokens           (tokens.go)
//
// The one thing not in the database is signing.key: a raw HS256 secret kept as a
// file under directory, generated on first use (signing.go). The mutex guards
// only that file; SQLite handles its own concurrency (WAL + busy_timeout), and
// the few read-modify-write flows use transactions.
package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
)

// Store is the persistence handle. Construct with New.
type Store struct {
	// directory holds signing.key; the database holds everything else.
	directory string
	database  *sql.DB
	// mu guards signing.key file access (signing.go) and serializes the TryPair
	// read-modify-write so concurrent "once" pairings can't both win.
	mu sync.Mutex
}

// New ensures directory exists (for signing.key) and returns a Store over the
// given database. The oauth_* schema is brought up by the db package's
// migrations before this is called; New only normalizes a legacy signing key.
func New(directory string, database *sql.DB) (*Store, error) {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	s := &Store{directory: directory, database: database}
	if err := s.migrateLegacySigningKey(); err != nil {
		return nil, err
	}
	return s, nil
}

// path joins directory + name (used for signing.key).
func (s *Store) path(name string) string { return filepath.Join(s.directory, name) }

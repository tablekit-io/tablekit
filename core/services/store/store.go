// Package store is the persistence layer for the OAuth/MCP server. It reads and
// writes the oauth_* tables in tablekit's Postgres database (schema owned by the
// db package's goose migrations).
//
// State that lives in Postgres:
//   - oauth_clients        registered clients + CLI bearer clients (clients.go)
//   - oauth_paired_clients  pairing set; config  pairing mode    (clients.go)
//   - oauth_auth_codes     one-time PKCE auth codes               (tokens.go)
//   - oauth_token_chains   refresh-token lineages                 (tokens.go)
//   - oauth_bearer_tokens  long-lived CLI bearer tokens           (tokens.go)
//
// The HS256 signing key is not persisted here: it is supplied via the SIGNING_KEY
// env and decoded by DecodeSigningKey (signing.go). Postgres handles its own
// concurrency; the mutex only serializes the TryPair read-modify-write.
package store

import (
	"database/sql"
	"sync"
)

// Store is the persistence handle. Construct with New.
type Store struct {
	database *sql.DB
	// mu serializes the TryPair read-modify-write so concurrent "once" pairings
	// can't both win.
	mu sync.Mutex
}

// New returns a Store over the given database. The oauth_* schema is brought up
// by the db package's migrations before this is called.
func New(database *sql.DB) *Store {
	return &Store{database: database}
}

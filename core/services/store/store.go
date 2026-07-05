// Package store is the persistence layer for the OAuth/MCP server: one repository
// per aggregate over the oauth_* tables in tablekit's Postgres database (schema
// owned by the db package's goose migrations). Queries are built with go-jet
// (typed table/model code under core/db/gen), so columns and types are checked at
// compile time rather than in hand-written SQL strings.
//
// The repositories:
//   - ClientRepository       oauth_clients (registered + CLI bearer clients)  (clients.go)
//   - AuthCodeRepository      oauth_auth_codes (one-time PKCE codes)           (tokens.go)
//   - TokenChainRepository    oauth_token_chains (refresh-token lineages)      (tokens.go)
//   - BearerTokenRepository   oauth_bearer_tokens (CLI bearer tokens)          (tokens.go)
//   - PairingRepository       oauth_paired_clients + config (pairing set/mode) (pairing.go)
//
// The HS256 signing key is not persisted here: it is supplied via the SIGNING_KEY
// env and decoded by DecodeSigningKey (signing.go). Postgres handles concurrency;
// PairingRepository additionally serializes its TryPair read-modify-write with a
// mutex so concurrent "once" pairings can't both win.
package store

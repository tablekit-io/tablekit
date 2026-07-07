// Package identity is the physical-database fingerprint the engine derives by
// connecting to a database. It is a stdlib-only leaf so drivers, the engine
// facade and the services layer can all speak in terms of Identity without an
// import cycle (drivers must not import services).
//
// A name in databases.yaml is only a label; the connection details behind it can
// be repointed at a completely different physical database. Identity pins the
// physical database itself: each driver derives a stable Key from server- and
// database-level identifiers (postgres system_identifier + database oid,
// mysql/mariadb server_uuid + schema) that survive restarts but differ across
// distinct databases. The static Key is the match key; Attributes carries the
// structured, per-family shape for observability.
package identity

import (
	"strings"
)

// Identity fingerprints one physical database.
type Identity struct {
	// Engine is the database family: "postgres", "mysql" or "mariadb".
	Engine string `json:"engine"`
	// Key is the stable, comparable form. Each driver owns its format, prefixed
	// by the engine so keys never collide across families.
	Key string `json:"key"`
	// Attributes is the structured fingerprint, persisted as jsonb for
	// observability. Its shape differs per family and is never matched on.
	Attributes map[string]string `json:"attributes"`
}

// Equal reports whether two identities name the same physical database. An empty
// Key never matches (a failed or partial derivation must not compare equal).
func (i Identity) Equal(other Identity) bool {
	return i.Key != "" && i.Key == other.Key
}

// Sanitize replaces every character outside [A-Za-z0-9._-] with '_', so a value
// interpolated into a Key can't introduce the '-' field separator or other
// surprises. Drivers use it when building their Key.
func Sanitize(value string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r == '.' || r == '_' || r == '-':
			return r
		default:
			return '_'
		}
	}, value)
}

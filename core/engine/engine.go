// Package engine loads database definitions and runs read-only SQL against them.
//
// The only thing consumers touch is Service: they name a database and a query,
// and get back a Result. Everything about how the connection is made — which
// driver library, whether an SSH tunnel or TLS is involved, how rows are
// decoded — lives in the subpackages (config, encoding, transport, driver) and
// is selected by the routing Service via the unexported databaseEngine interface.
package engine

import (
	"context"
	"fmt"
	"sort"

	"core/engine/config"
	"core/engine/driver/mysql"
	"core/engine/driver/postgres"
	"core/engine/encoding"
)

// Limits, Result and OmittedColumn are surfaced from the subpackages that own
// them so consumers depend only on engine.
type (
	Limits        = config.Limits
	Result        = encoding.Result
	OmittedColumn = encoding.OmittedColumn
)

// DatabaseInfo is the public, secret-free description of a configured database.
type DatabaseInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// databaseEngine is one database-engine implementation. Each implementation
// owns its driver, SSH tunnelling, TLS, pagination and value normalization
// entirely; none of those types appear in this signature, so consumers stay
// decoupled from them. Run applies the page window in the engine's own SQL
// dialect, over-fetches to detect more rows, and reports hasMore.
type databaseEngine interface {
	Run(ctx context.Context, db config.Database, query string, page config.Page, limits config.Limits) (result *encoding.Result, hasMore bool, err error)
}

// engineFor routes a database type to its implementation.
func engineFor(dbType config.DatabaseType) (databaseEngine, error) {
	switch dbType {
	case config.DatabaseTypePostgres:
		return postgres.Engine{}, nil
	case config.DatabaseTypeMySQL:
		return mysql.NewMySQL(), nil
	case config.DatabaseTypeMariaDB:
		return mysql.NewMariaDB(), nil
	default:
		return nil, fmt.Errorf("unsupported database type %q", dbType)
	}
}

// Service holds the resolved databases and the query limits. It is immutable
// after Load, so it is safe to share and call concurrently.
type Service struct {
	databases map[string]config.Database
	limits    config.Limits
}

// Load reads and validates the databases YAML at path, resolves secrets and
// defaults, and returns a ready Service. A missing file is tolerated and yields
// a Service with no databases.
func Load(path string, limits Limits) (*Service, error) {
	databases, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	return &Service{databases: databases, limits: limits.WithDefaults()}, nil
}

// RunReadOnly runs query against the named database inside a read-only
// transaction and returns the (possibly truncated) result. It routes to the
// engine implementation for the database's type; the caller never learns which
// driver, tunnel or TLS settings were used.
func (s *Service) RunReadOnly(ctx context.Context, databaseName, query string) (*Result, error) {
	db, ok := s.databases[databaseName]
	if !ok {
		return nil, fmt.Errorf("unknown database %q", databaseName)
	}
	implementation, err := engineFor(db.Type)
	if err != nil {
		return nil, err
	}
	result, _, err := implementation.Run(ctx, db, query, config.Page{Limit: s.limits.MaxRows}, s.limits)
	return result, err
}

// PageOptions tunes a single paginated run. Any zero field falls back to a
// default: Limit to the service's MaxRows, MaxBytes to the service's MaxBytes.
type PageOptions struct {
	// Offset is the number of leading rows to skip (the SQL OFFSET).
	Offset int
	// Limit is the maximum number of rows to return in this window.
	Limit int
	// MaxBytes caps the encoded size of the returned rows; charts/exports raise
	// it well above the run_query default.
	MaxBytes int
}

// RunReadOnlyPage runs query against the named database as a paginated window and
// reports whether more rows exist beyond it. It wraps the query in a LIMIT/OFFSET
// subquery and over-fetches one row so hasMore can be detected without a second
// round-trip; the extra row is trimmed before returning. Like RunReadOnly it
// executes inside a read-only transaction under the statement timeout.
func (s *Service) RunReadOnlyPage(ctx context.Context, databaseName, query string, opts PageOptions) (result *Result, hasMore bool, err error) {
	db, ok := s.databases[databaseName]
	if !ok {
		return nil, false, fmt.Errorf("unknown database %q", databaseName)
	}
	implementation, err := engineFor(db.Type)
	if err != nil {
		return nil, false, err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = s.limits.MaxRows
	}
	maxBytes := opts.MaxBytes
	if maxBytes <= 0 {
		maxBytes = s.limits.MaxBytes
	}

	// The driver owns the page window: it applies opts.Offset/limit in its own SQL
	// dialect, over-fetches to detect hasMore, and caps at limit+1 itself. Only the
	// timeout and byte cap flow through Limits.
	limits := config.Limits{
		StatementTimeout: s.limits.StatementTimeout,
		MaxBytes:         maxBytes,
	}
	page := config.Page{Offset: opts.Offset, Limit: limit}
	return implementation.Run(ctx, db, query, page, limits)
}

// List returns the configured databases, name and type only, sorted by name.
func (s *Service) List() []DatabaseInfo {
	infos := make([]DatabaseInfo, 0, len(s.databases))
	for name, db := range s.databases {
		infos = append(infos, DatabaseInfo{Name: name, Type: string(db.Type)})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos
}

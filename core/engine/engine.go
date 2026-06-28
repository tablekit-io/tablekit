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
// owns its driver, SSH tunnelling and TLS entirely; none of those types appear
// in this signature, so consumers stay decoupled from them.
type databaseEngine interface {
	Run(ctx context.Context, db config.Database, query string, limits config.Limits) (*encoding.Result, error)
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
	return implementation.Run(ctx, db, query, s.limits)
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

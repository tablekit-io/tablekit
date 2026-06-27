package engine

import (
	"context"
	"fmt"
	"sort"
)

// RunReadOnly runs query against the named database inside a read-only
// transaction and returns the (possibly truncated) result. It routes to the
// engine implementation for the database's type; the caller never learns which
// driver, tunnel or TLS settings were used.
func (s *Service) RunReadOnly(ctx context.Context, databaseName, query string) (*Result, error) {
	db, ok := s.databases[databaseName]
	if !ok {
		return nil, fmt.Errorf("unknown database %q", databaseName)
	}
	implementation, err := engineFor(db.dbType)
	if err != nil {
		return nil, err
	}
	return implementation.run(ctx, db, query, s.limits)
}

// List returns the configured databases, name and type only, sorted by name.
func (s *Service) List() []DatabaseInfo {
	infos := make([]DatabaseInfo, 0, len(s.databases))
	for name, db := range s.databases {
		infos = append(infos, DatabaseInfo{Name: name, Type: string(db.dbType)})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos
}

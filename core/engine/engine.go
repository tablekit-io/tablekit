// Package engine loads database definitions and runs read-only SQL against them.
//
// The only thing consumers touch is Service: they name a database and a query,
// and get back a Result. Everything about how the connection is made — which
// driver library, whether an SSH tunnel or TLS is involved, how rows are
// decoded — is hidden behind a per-engine implementation of the unexported
// databaseEngine interface, selected by the routing Service.
package engine

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed schemas/databases.schema.json
var schemaJSON []byte

// Limits bound a single query: a server-side statement timeout plus caps on the
// number of rows and the encoded byte size of the result. The first cap reached
// truncates the result and sets Result.Truncated.
type Limits struct {
	StatementTimeout time.Duration
	MaxRows          int
	MaxBytes         int
}

func (l Limits) withDefaults() Limits {
	if l.StatementTimeout <= 0 {
		l.StatementTimeout = 10 * time.Second
	}
	if l.MaxRows <= 0 {
		l.MaxRows = 2048
	}
	if l.MaxBytes <= 0 {
		l.MaxBytes = 64 * 1024
	}
	return l
}

// Result is a query's outcome: the column names, the rows as objects, and flags
// telling the caller when the result was cut short or had columns dropped.
type Result struct {
	Columns   []string         `json:"columns"`
	Rows      []map[string]any `json:"rows"`
	RowCount  int              `json:"row_count"`
	Truncated bool             `json:"truncated"`
	// Omitted lists columns whose values could not be represented as JSON text
	// (e.g. binary/non-UTF-8) and were dropped from Rows, so the caller knows
	// the column exists but was excluded, and why.
	Omitted []OmittedColumn `json:"omitted,omitempty"`
}

// OmittedColumn names a dropped column and the reason it was dropped.
type OmittedColumn struct {
	Column string `json:"column"`
	Reason string `json:"reason"`
}

// DatabaseInfo is the public, secret-free description of a configured database.
type DatabaseInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// databaseEngine is one database-engine implementation. Each implementation
// owns its driver, SSH tunnelling and TLS entirely; none of those types appear
// in this signature, so consumers stay decoupled from them.
type databaseEngine interface {
	run(ctx context.Context, db database, query string, limits Limits) (*Result, error)
}

// engineFor routes a database type to its implementation.
func engineFor(dbType databaseType) (databaseEngine, error) {
	switch dbType {
	case databaseTypePostgres:
		return postgresEngine{}, nil
	case databaseTypeMySQL:
		return newMySQLEngine(), nil
	case databaseTypeMariaDB:
		return newMariaDBEngine(), nil
	default:
		return nil, fmt.Errorf("unsupported database type %q", dbType)
	}
}

// Service holds the resolved databases and the query limits. It is immutable
// after Load, so it is safe to share and call concurrently.
type Service struct {
	databases map[string]database
	limits    Limits
}

// Load reads and validates the databases YAML at path against the embedded JSON
// Schema, resolves secrets and defaults, and returns a ready Service. A missing
// file is tolerated and yields a Service with no databases.
func Load(path string, limits Limits) (*Service, error) {
	limits = limits.withDefaults()

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &Service{databases: map[string]database{}, limits: limits}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read databases file %q: %w", path, err)
	}

	var instance any
	if err := yaml.Unmarshal(data, &instance); err != nil {
		return nil, fmt.Errorf("parse databases file %q: %w", path, err)
	}
	if err := validate(instance); err != nil {
		return nil, fmt.Errorf("databases config %q is invalid: %w", path, err)
	}

	var file databasesFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode databases file %q: %w", path, err)
	}

	databases := make(map[string]database, len(file.Databases))
	for name, raw := range file.Databases {
		resolved, err := raw.resolve(name)
		if err != nil {
			return nil, err
		}
		databases[name] = resolved
	}
	return &Service{databases: databases, limits: limits}, nil
}

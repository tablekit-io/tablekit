// Package shared holds the dependencies, helpers and schema plumbing the
// per-tool MCP handler packages build on. Each tool lives in its own package
// (one directory per tool) and depends on this one; the parent handlers package
// wires them onto an mcp.Server.
package shared

import (
	"context"
	"fmt"

	"core/engine"
	"core/services/databases"
	"core/services/oauth"
	"core/services/queries"

	"github.com/google/uuid"
)

// Deps carries the dependencies the tools need: the query engine, the
// stored-query repository, the physical-database identity resolver, the JWT
// issuer (for signed export URLs) and the public base URL those URLs are built
// against.
type Deps struct {
	Engine        *engine.Service
	Queries       queries.QueryRepository
	Databases     *databases.Resolver
	Issuer        *oauth.Issuer
	PublicBaseURL string
}

// RequireQuery confirms a stored query exists for key, turning an unknown key
// into a user-facing error. A malformed key (not a UUID) can't identify any
// stored query, so it is reported the same as an unknown one.
func (d Deps) RequireQuery(ctx context.Context, key string) error {
	id, err := uuid.Parse(key)
	if err != nil {
		return fmt.Errorf("unknown query_key %q (run query_database first)", key)
	}
	descriptor, err := d.Queries.Get(ctx, id)
	if err != nil {
		return err
	}
	if descriptor == nil {
		return fmt.Errorf("unknown query_key %q (run query_database first)", key)
	}
	return nil
}

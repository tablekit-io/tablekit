// Package shared holds the dependencies, helpers and schema plumbing the
// per-tool MCP handler packages build on. Each tool lives in its own package
// (one directory per tool) and depends on this one; the parent handlers package
// wires them onto an mcp.Server.
package shared

import (
	"context"
	"fmt"

	"core/engine"
	"core/services/oauth"
	"core/services/queries"
)

// Deps carries the dependencies the tools need: the query engine, the
// stored-query repository, the JWT issuer (for signed export URLs) and the
// public base URL those URLs are built against.
type Deps struct {
	Engine        *engine.Service
	Queries       queries.QueryRepository
	Issuer        *oauth.Issuer
	PublicBaseURL string
}

// RequireQuery confirms a stored query exists for key, turning an unknown key
// into a user-facing error.
func (d Deps) RequireQuery(ctx context.Context, key string) error {
	descriptor, err := d.Queries.Get(ctx, key)
	if err != nil {
		return err
	}
	if descriptor == nil {
		return fmt.Errorf("unknown query_key %q (run query_database first)", key)
	}
	return nil
}

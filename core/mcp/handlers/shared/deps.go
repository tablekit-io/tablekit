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

// Deps is the full dependency set for the MCP tools: the query engine, the
// stored-query repository, the physical-database identity resolver, the JWT
// issuer (for signed export URLs) and the public base URL those URLs are built
// against. It is the aggregate the parent handlers package constructs and wires;
// individual tools take only the narrow slice they need (Query, Export or the
// QueryGuard interface) rather than the whole bundle.
type Deps struct {
	Engine        *engine.Service
	Queries       queries.QueryRepository
	Databases     *databases.Resolver
	Issuer        *oauth.Issuer
	PublicBaseURL string
}

// QueryDeps is what the query-execution tools (query_database, read_results,
// fetch_chart_data) need: run SQL against a resolved physical database and
// save/load the stored query descriptor.
type QueryDeps struct {
	Engine    *engine.Service
	Queries   queries.QueryRepository
	Databases *databases.Resolver
}

// ExportDeps is what get_export_url needs: verify the stored query's database,
// then mint a signed download URL against the public base URL.
type ExportDeps struct {
	Queries       queries.QueryRepository
	Databases     *databases.Resolver
	Issuer        *oauth.Issuer
	PublicBaseURL string
}

// QueryGuard is the single capability the chart tools need: confirm a stored
// query exists before rendering a widget for it. Deps satisfies it.
type QueryGuard interface {
	RequireQuery(ctx context.Context, key string) error
}

// Query returns the query-execution slice of the dependencies.
func (d Deps) Query() QueryDeps {
	return QueryDeps{Engine: d.Engine, Queries: d.Queries, Databases: d.Databases}
}

// Export returns the export-signing slice of the dependencies.
func (d Deps) Export() ExportDeps {
	return ExportDeps{Queries: d.Queries, Databases: d.Databases, Issuer: d.Issuer, PublicBaseURL: d.PublicBaseURL}
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

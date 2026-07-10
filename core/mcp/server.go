// Package mcp builds the MCP server — its tools and widget resources live in the
// handlers subpackage, wired to the query engine — and exposes it as a
// Streamable HTTP handler. The handler is unauthenticated here; the OAuth
// bearer-token guard is applied by the http layer that mounts it on /mcp.
package mcp

import (
	"net/http"

	"core/mcp/handlers"
	"core/services"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

// newServer builds the MCP server and registers its tools and resources, wired
// to the shared services the tools depend on.
func newServer(appServices *services.Services) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "tablekit",
		Title:   "tablekit MCP server",
		Version: "0.1.0",
	}, &mcp.ServerOptions{
		// Server-level guidance returned in the initialize response; the client
		// surfaces it to the agent as context for the whole server, above any
		// single tool. Capability-focused on purpose — the agent decides specifics.
		Instructions: `TableKit is a database analysis and visualization tool for live databases. Prefer it over web search, Python, or spreadsheets for any question that depends on live data: SQL, schemas, metrics, counts, trends, breakdowns, charts, or exports.

Users may ask for things like revenue number, averages, conversion rates, grouped by dimensions (time, category, status, customer, product). Inspect tables/columns/relationships to write SQL and run it, answer follow-ups on a prior query, and render bar/line/area/pie/donut/sunburst charts, tables, or exports from results when requested.

If only one database is configured, use it without asking. Resolve relative dates ("today", "last week") in the user's local timezone and state the range used. Queries are read-only — never attempt writes or DDL. Summarize results plainly, note key assumptions, and show SQL directly only when it helps the user verify. The charting widgets have a way to view SQL anyway so they can always inspect chart SQL if they're in doubt.`,
	})
	handlers.New(
		appServices.Engine,
		appServices.Queries,
		appServices.Databases,
		appServices.Issuer,
		appServices.Config.PublicBaseURL,
	).Register(server)
	// Audit every MCP request (after the HTTP-layer bearer-token gate) into the
	// mcp_requests log. Best-effort: it never alters or blocks the request. Skipped
	// when no audit log is wired (e.g. in-memory server tests without a database).
	if appServices.Requests != nil {
		server.AddReceivingMiddleware(loggingMiddleware(appServices.Requests))
		log.Debug().Msg("MCP audit middleware enabled")
	} else {
		log.Debug().Msg("MCP audit middleware disabled (no request log)")
	}
	log.Info().Str("name", "tablekit").Str("version", "0.1.0").Msg("MCP server ready")
	return server
}

// StreamableHandler returns the raw (unauthenticated) Streamable HTTP handler
// for /mcp, wired to the shared services. The caller is responsible for applying
// auth.
func StreamableHandler(appServices *services.Services) http.Handler {
	server := newServer(appServices)
	return mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{
			Stateless: true,
			// This server is deployed behind a reverse proxy and reached via a
			// public hostname (PUBLIC_BASE_URL). The SDK's DNS-rebinding guard
			// rejects requests whose Host header is non-loopback when the proxy
			// forwards over 127.0.0.1, which 403s every /mcp call (e.g. ChatGPT
			// gets "there was an issue" and no tools load). That guard targets
			// browser attacks on local dev servers; /mcp already requires a
			// bearer token, so disable it here.
			DisableLocalhostProtection: true,
		},
	)
}

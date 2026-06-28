// Package mcp builds the MCP server — its tools and widget resources live in the
// handlers subpackage, wired to the query engine — and exposes it as a
// Streamable HTTP handler. The handler is unauthenticated here; the OAuth
// bearer-token guard is applied by the http layer that mounts it on /mcp.
package mcp

import (
	"net/http"

	"core/engine"
	"core/mcp/handlers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newServer builds the MCP server and registers its tools and resources, wired
// to the query engine.
func newServer(engineService *engine.Service) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "tablekit",
		Title:   "tablekit MCP server",
		Version: "0.1.0",
	}, nil)
	handlers.New(engineService).Register(server)
	return server
}

// StreamableHandler returns the raw (unauthenticated) Streamable HTTP handler
// for /mcp, wired to the query engine. The caller is responsible for applying
// auth.
func StreamableHandler(engineService *engine.Service) http.Handler {
	server := newServer(engineService)
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

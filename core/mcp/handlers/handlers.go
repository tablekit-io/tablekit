// Package handlers holds the MCP tools this server exposes, one tool per file,
// plus the registration that wires them (and the built widget UI resources)
// onto an mcp.Server.
package handlers

import (
	"context"

	"core/engine"
	"core/mcp/ui"
	"core/services/oauth"
	"core/services/queries"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Handlers serves the MCP tools. It carries the dependencies the tools need: the
// query engine, the stored-query repository, the JWT issuer (for signed export
// URLs) and the public base URL those URLs are built against. Construct with New
// and wire with Register.
type Handlers struct {
	Engine        *engine.Service
	Queries       queries.QueryRepository
	Issuer        *oauth.Issuer
	PublicBaseURL string
}

// New wires the MCP tool handlers to their dependencies.
func New(engineService *engine.Service, queriesRepo queries.QueryRepository, issuer *oauth.Issuer, publicBaseURL string) *Handlers {
	return &Handlers{
		Engine:        engineService,
		Queries:       queriesRepo,
		Issuer:        issuer,
		PublicBaseURL: publicBaseURL,
	}
}

// Register wires every tool and the built widget UI resources onto s.
func (h *Handlers) Register(s *mcp.Server) {
	h.registerListAvailableDatabases(s)
	h.registerQueryDatabase(s)
	h.registerReadResults(s)
	h.registerFetchChartData(s)
	h.registerShowBarLineAreaChart(s)
	h.registerShowPieDonutSunburstChart(s)
	h.registerGetExportURL(s)
	registerWidgetResources(s)
}

// registerWidgetResources registers the built widget templates as ui:// resources
// the host can render in a sandboxed iframe. Empty until @tablekit/widgets is built.
func registerWidgetResources(s *mcp.Server) {
	for _, resource := range ui.Resources() {
		uri := resource.URI
		mime := resource.MIMEType
		html := resource.HTML
		s.AddResource(
			&mcp.Resource{Name: resource.Name, URI: uri, MIMEType: mime},
			func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{URI: uri, MIMEType: mime, Text: html},
					},
				}, nil
			},
		)
	}
}

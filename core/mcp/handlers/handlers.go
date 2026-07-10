// Package handlers wires the MCP tools this server exposes — one package per
// tool under this directory — plus the built widget UI resources, onto an
// mcp.Server. The tools share dependencies and helpers via the shared package.
package handlers

import (
	"context"

	"core/engine"
	fetchchartdata "core/mcp/handlers/fetch_chart_data"
	getexporturl "core/mcp/handlers/get_export_url"
	listavailabledatabases "core/mcp/handlers/list_available_databases"
	querydatabase "core/mcp/handlers/query_database"
	readresults "core/mcp/handlers/read_results"
	"core/mcp/handlers/shared"
	showbarlineareachart "core/mcp/handlers/show_bar_line_area_chart"
	showpiedonutsunburstchart "core/mcp/handlers/show_pie_donut_sunburst_chart"
	"core/mcp/ui"
	"core/services/databases"
	"core/services/oauth"
	"core/services/queries"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

// Handlers serves the MCP tools. It carries the dependencies the tools need: the
// query engine, the stored-query repository, the JWT issuer (for signed export
// URLs) and the public base URL those URLs are built against. Construct with New
// and wire with Register.
type Handlers struct {
	deps shared.Deps
}

// New wires the MCP tool handlers to their dependencies.
func New(engineService *engine.Service, queriesRepo queries.QueryRepository, resolver *databases.Resolver, issuer *oauth.Issuer, publicBaseURL string) *Handlers {
	return &Handlers{deps: shared.Deps{
		Engine:        engineService,
		Queries:       queriesRepo,
		Databases:     resolver,
		Issuer:        issuer,
		PublicBaseURL: publicBaseURL,
	}}
}

// Register wires every tool and the built widget UI resources onto s.
func (h *Handlers) Register(s *mcp.Server) {
	listavailabledatabases.Register(s, h.deps)
	querydatabase.Register(s, h.deps)
	readresults.Register(s, h.deps)
	fetchchartdata.Register(s, h.deps)
	showbarlineareachart.Register(s, h.deps)
	showpiedonutsunburstchart.Register(s, h.deps)
	getexporturl.Register(s, h.deps)
	registerWidgetResources(s)
	log.Info().Int("tools", 7).Msg("MCP tools registered")
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

// Package handlers holds the MCP tools this server exposes, one tool per file,
// plus the registration that wires them (and the built widget UI resources)
// onto an mcp.Server.
package handlers

import (
	"context"

	"core/mcp/ui"
	"core/services"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Handlers serves the MCP tools. It carries the shared Services so tools can
// reach config and the store. Construct with New and wire with Register.
type Handlers struct {
	Services *services.Services
}

// New wires the MCP tool handlers to the shared services.
func New(appServices *services.Services) *Handlers {
	return &Handlers{Services: appServices}
}

// Register wires every tool and the built widget UI resources onto s.
func (h *Handlers) Register(s *mcp.Server) {
	h.registerHelloWorld(s)
	h.registerHelloWorldInteractive(s)
	h.registerHelloWorldInteractiveData(s)
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

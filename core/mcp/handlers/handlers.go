// Package handlers holds the MCP tools this server exposes, one tool per file,
// plus the registration that wires them (and the built widget UI resources)
// onto an mcp.Server.
package handlers

import (
	"context"

	"core/mcp/ui"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// pointer returns a pointer to v, for the optional *bool annotation hints.
func pointer[T any](v T) *T { return &v }

// Register wires every tool and the built widget UI resources onto s.
func Register(s *mcp.Server) {
	registerHelloWorld(s)
	registerHelloWorldInteractive(s)
	registerHelloWorldInteractiveData(s)
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

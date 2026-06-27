// Package app wires the public-facing engine: the OAuth 2.1 endpoints, the MCP
// route, and the welcome root. RegisterHandlers is the single entry point the
// http layer calls to mount everything on the app engine.
package app

import (
	"core/http/app/oauth"
	"core/http/control"
	"core/services"

	"github.com/gin-gonic/gin"
)

// RegisterHandlers mounts the whole app engine — OAuth, MCP, and the welcome
// root — on engine, wired to the shared services. It errors only if the OAuth
// layer (persistence / signing key) fails to initialize.
func RegisterHandlers(engine *gin.Engine, appServices *services.Services) error {
	oauthHandlers, err := oauth.NewHandlers(appServices)
	if err != nil {
		return err
	}

	engine.GET("/", control.Welcome("hello and welcome to the tablekit MCP server"))
	oauthHandlers.Register(engine)

	// The SDK's bearer middleware + streamable handler are net/http; gin.WrapH
	// adapts them. /mcp must accept GET, POST and DELETE.
	engine.Any("/mcp", gin.WrapH(MCPRoute(appServices, oauthHandlers.Issuer())))

	return nil
}

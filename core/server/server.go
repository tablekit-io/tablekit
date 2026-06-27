// Package server builds the two Gin engines this service runs:
//
//   - app engine:     MCP (/mcp) + OAuth (/oauth/*, /register, /.well-known/*)
//   - control engine: /, /health (liveness; reserved for future ops)
package server

import (
	"core/mcpserver"
	"core/oauth"
	"core/server/handlers"
	"core/services"

	"github.com/gin-gonic/gin"
)

// App holds the constructed engines and the services that drive the listeners.
type App struct {
	Services *services.Services
	AppEng   *gin.Engine
	Control  *gin.Engine
}

// Build wires storage, OAuth and MCP into the two engines. It returns an error
// if persistence or the signing key cannot be initialized.
func Build(appServices *services.Services) (*App, error) {
	oauthHandlers, err := oauth.NewHandlers(appServices)
	if err != nil {
		return nil, err
	}

	appEng := gin.New()
	appEng.Use(gin.Logger(), gin.Recovery())
	appEng.GET("/", handlers.Welcome("hello and welcome to the tablekit MCP server"))
	oauthHandlers.Register(appEng)
	// The SDK's bearer middleware + streamable handler are net/http; gin.WrapH
	// adapts them. /mcp must accept GET, POST and DELETE.
	mcpHandler := mcpserver.Handler(appServices, oauthHandlers.Issuer())
	appEng.Any("/mcp", gin.WrapH(mcpHandler))

	control := gin.New()
	control.Use(gin.Logger(), gin.Recovery())
	control.GET("/", handlers.Welcome("hello and welcome to the tablekit control server"))
	control.GET("/health", handlers.Health)

	return &App{Services: appServices, AppEng: appEng, Control: control}, nil
}

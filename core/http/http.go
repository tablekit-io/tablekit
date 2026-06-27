// Package http builds the two Gin engines this service runs:
//
//   - app engine:     MCP (/mcp) + OAuth (/oauth/*, /register, /.well-known/*)
//   - control engine: /, /health (liveness; reserved for future ops)
package http

import (
	"core/http/app/mcp"
	"core/http/app/oauth"
	"core/http/control"
	"core/services"

	"github.com/gin-gonic/gin"
)

// App holds the constructed engines and the services that drive the listeners.
type App struct {
	Services      *services.Services
	AppEngine     *gin.Engine
	ControlEngine *gin.Engine
}

// Build wires storage, OAuth and MCP into the two engines. It returns an error
// if persistence or the signing key cannot be initialized.
func Build(appServices *services.Services) (*App, error) {
	oauthHandlers, err := oauth.NewHandlers(appServices)
	if err != nil {
		return nil, err
	}

	appEngine := gin.New()
	appEngine.Use(gin.Logger(), gin.Recovery())
	appEngine.GET("/", control.Welcome("hello and welcome to the tablekit MCP server"))
	oauthHandlers.Register(appEngine)

	// The SDK's bearer middleware + streamable handler are net/http; gin.WrapH
	// adapts them. /mcp must accept GET, POST and DELETE.
	mcpHandler := mcp.Handler(appServices, oauthHandlers.Issuer())
	appEngine.Any("/mcp", gin.WrapH(mcpHandler))

	controlEngine := gin.New()
	controlEngine.Use(gin.Logger(), gin.Recovery())
	controlEngine.GET("/", control.Welcome("hello and welcome to the tablekit control server"))
	controlEngine.GET("/health", control.Health)

	return &App{Services: appServices, AppEngine: appEngine, ControlEngine: controlEngine}, nil
}

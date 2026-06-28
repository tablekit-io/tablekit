// Package http builds the two Gin engines this service runs:
//
//   - app engine:     MCP (/mcp) + OAuth (/oauth/*, /register, /.well-known/*)
//   - control engine: /, /health (liveness; reserved for future ops)
package http

import (
	"core/http/app"
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

// Build constructs the two engines and hands each its registration function and
// the shared services.
func Build(appServices *services.Services) *App {
	appEngine := gin.New()
	appEngine.Use(gin.Logger(), gin.Recovery())
	app.RegisterHandlers(appEngine, appServices)

	controlEngine := gin.New()
	controlEngine.Use(gin.Logger(), gin.Recovery())
	control.RegisterHandlers(controlEngine, appServices)

	return &App{Services: appServices, AppEngine: appEngine, ControlEngine: controlEngine}
}

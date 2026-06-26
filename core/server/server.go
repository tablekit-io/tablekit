// Package server builds the two Gin engines this service runs:
//
//   - app engine:     MCP (/mcp) + OAuth (/oauth/*, /register, /.well-known/*)
//   - control engine: /, /health (liveness; reserved for future ops)
package server

import (
	"core/config"
	"core/mcpserver"
	"core/oauth"
	"core/server/handlers"
	"core/store"

	"github.com/gin-gonic/gin"
)

// App holds the constructed engines and the config that drives the listeners.
type App struct {
	Cfg     *config.Config
	AppEng  *gin.Engine
	Control *gin.Engine
}

// Build wires storage, OAuth and MCP into the two engines. It returns an error
// if persistence or the signing key cannot be initialized.
func Build(cfg *config.Config) (*App, error) {
	st, err := store.New(cfg.DataDir)
	if err != nil {
		return nil, err
	}

	oauthHandlers, err := oauth.NewHandlers(cfg, st)
	if err != nil {
		return nil, err
	}

	appEng := gin.New()
	appEng.Use(gin.Logger(), gin.Recovery())
	appEng.GET("/", handlers.Welcome("hello and welcome to the tablekit MCP server"))
	oauthHandlers.Register(appEng)
	// The SDK's bearer middleware + streamable handler are net/http; gin.WrapH
	// adapts them. /mcp must accept GET, POST and DELETE.
	mcpHandler := mcpserver.Handler(cfg, oauthHandlers.Issuer())
	appEng.Any("/mcp", gin.WrapH(mcpHandler))

	control := gin.New()
	control.Use(gin.Logger(), gin.Recovery())
	control.GET("/", handlers.Welcome("hello and welcome to the tablekit control server"))
	control.GET("/health", handlers.Health)

	return &App{Cfg: cfg, AppEng: appEng, Control: control}, nil
}

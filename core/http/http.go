// Package http builds the two Gin engines this service runs:
//
//   - app engine:     MCP (/mcp) + OAuth (/oauth/*, /register, /.well-known/*)
//   - control engine: /, /health (liveness; reserved for future ops)
package http

import (
	"os"
	"time"

	"core/http/app"
	"core/http/control"
	"core/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func init() {
	// Run Gin in release mode outside development so it does not dump its raw
	// [GIN-debug] route/warning lines at startup. Mirrors config.IsDevelopment's
	// exact-match semantics; read from the environment here to avoid coupling this
	// package init to services/config load order.
	if os.Getenv("TABLEKIT_ENV") != "development" {
		gin.SetMode(gin.ReleaseMode)
	}
	// Point Gin's own writers at zerolog so any internal prints it still makes
	// (dev-mode debug lines, panic recovery) come out as JSON, not raw text.
	gin.DefaultWriter = log.Logger
	gin.DefaultErrorWriter = log.Logger
}

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
	appEngine.Use(requestLogger(), gin.Recovery())
	app.RegisterHandlers(appEngine, appServices)

	controlEngine := gin.New()
	controlEngine.Use(requestLogger(), gin.Recovery())
	control.RegisterHandlers(controlEngine, appServices)

	return &App{Services: appServices, AppEngine: appEngine, ControlEngine: controlEngine}
}

// requestLogger is a Gin middleware that emits one structured zerolog event per
// request, replacing gin.Logger's raw [GIN] access lines. The event level tracks
// the response status so error responses surface at error/warn level.
func requestLogger() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()

		context.Next()

		status := context.Writer.Status()
		event := log.Info()
		switch {
		case status >= 500:
			event = log.Error()
		case status >= 400:
			event = log.Warn()
		}

		path := context.Request.URL.Path
		if rawQuery := context.Request.URL.RawQuery; rawQuery != "" {
			path = path + "?" + rawQuery
		}

		event.
			Int("status", status).
			Str("method", context.Request.Method).
			Str("path", path).
			Str("client_ip", context.ClientIP()).
			Int64("latency_ms", time.Since(start).Milliseconds()).
			Int("size", context.Writer.Size()).
			Msg("request")
	}
}

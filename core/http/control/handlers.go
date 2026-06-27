package control

import (
	"core/http/commons"
	"core/services"

	"github.com/gin-gonic/gin"
)

// RegisterHandlers mounts the control engine — the welcome root and the health
// check — on engine. appServices is accepted for symmetry with the app engine
// and future ops checks (e.g. probing the store).
func RegisterHandlers(engine *gin.Engine, appServices *services.Services) {
	engine.GET("/", commons.Welcome("hello and welcome to the tablekit control server"))
	engine.GET("/health", Health)
}

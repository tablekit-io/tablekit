package commons

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Welcome returns a handler that serves a JSON hello/welcome message at the
// root path. The message identifies which listener (app vs control) answered.
func Welcome(message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   message,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

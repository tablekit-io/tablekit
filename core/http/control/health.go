package control

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Health responds with service status and the current UTC timestamp.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

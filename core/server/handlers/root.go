package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Root responds with a plain-text greeting.
func Root(c *gin.Context) {
	c.String(http.StatusOK, "hello world")
}

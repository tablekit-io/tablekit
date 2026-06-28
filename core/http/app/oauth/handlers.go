package oauth

import (
	"net/http"
	"sync"

	"core/services"

	"github.com/gin-gonic/gin"
)

// Handlers serves the OAuth 2.1 endpoints. Construct with NewHandlers and mount
// with Register. The JWT issuer lives on the shared services
// (appServices.Issuer); these handlers just read it.
type Handlers struct {
	appServices *services.Services
	// authorizeMu serializes the whole /authorize handler so the
	// read-client → pair → issue-code sequence is atomic and two concurrent
	// requests can never both win a pairing slot.
	authorizeMu sync.Mutex
}

// NewHandlers wires the OAuth layer to its services.
func NewHandlers(appServices *services.Services) *Handlers {
	return &Handlers{appServices: appServices}
}

// sendError writes an RFC 6749 error response with no-store caching.
func sendError(c *gin.Context, status int, code, description string) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	body := gin.H{"error": code}
	if description != "" {
		body["error_description"] = description
	}
	c.JSON(status, body)
}

// noStoreJSON writes a success payload with no-store caching (token endpoint).
func noStoreJSON(c *gin.Context, body any) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, body)
}

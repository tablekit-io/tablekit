package oauth

import (
	"net/http"

	"core/config"
	"core/store"

	"github.com/gin-gonic/gin"
)

// Handlers serves the OAuth 2.1 endpoints. Construct with NewHandlers and mount
// with Register.
type Handlers struct {
	cfg    *config.Config
	store  *store.Store
	issuer *Issuer
}

// NewHandlers wires the OAuth layer to its config, persistence and JWT issuer.
func NewHandlers(cfg *config.Config, st *store.Store) (*Handlers, error) {
	issuer, err := NewIssuer(cfg, st)
	if err != nil {
		return nil, err
	}
	return &Handlers{cfg: cfg, store: st, issuer: issuer}, nil
}

// Issuer exposes the JWT issuer so the MCP layer can verify access tokens.
func (h *Handlers) Issuer() *Issuer { return h.issuer }

// Register mounts every OAuth route on the given Gin engine.
func (h *Handlers) Register(r *gin.Engine) {
	r.POST("/register", h.handleRegister)
	r.GET("/oauth/authorize", h.handleAuthorize)
	r.POST("/oauth/token", h.handleToken)
	r.GET("/.well-known/oauth-authorization-server", h.handleAuthServerMetadata)
	r.GET("/.well-known/oauth-protected-resource", h.handleProtectedResourceMetadata)
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

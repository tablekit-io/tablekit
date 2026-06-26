package oauth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleAuthServerMetadata serves RFC 8414 authorization-server metadata so
// clients can discover the authorize/token/register endpoints.
func (h *Handlers) handleAuthServerMetadata(c *gin.Context) {
	base := h.cfg.PublicBaseURL
	c.JSON(http.StatusOK, gin.H{
		"issuer":                                base,
		"authorization_endpoint":                base + "/oauth/authorize",
		"token_endpoint":                        base + "/oauth/token",
		"registration_endpoint":                 base + "/register",
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"response_types_supported":              []string{"code"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"scopes_supported":                      []string{Scope},
	})
}

// handleProtectedResourceMetadata serves RFC 9728 protected-resource metadata.
// Since the resource and authorization servers share one origin, both point at
// PublicBaseURL.
func (h *Handlers) handleProtectedResourceMetadata(c *gin.Context) {
	base := h.cfg.PublicBaseURL
	c.JSON(http.StatusOK, gin.H{
		"resource":                 base,
		"authorization_servers":    []string{base},
		"bearer_methods_supported": []string{"header"},
		"scopes_supported":         []string{Scope},
	})
}

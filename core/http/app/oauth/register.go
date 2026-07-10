package oauth

import (
	"net/http"
	"time"

	"core/services/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// registerRequest is the subset of RFC 7591 dynamic client registration we
// honor. MCP clients POST this to obtain a client_id before /authorize.
type registerRequest struct {
	RedirectURIs []string `json:"redirect_uris"`
	ClientName   string   `json:"client_name"`
}

// HandleRegister implements POST /register (RFC 7591). It accepts any client —
// the single-client pairing lock is enforced later, at /authorize.
func (h *Handlers) HandleRegister(c *gin.Context) {
	var request registerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		sendError(c, http.StatusBadRequest, "invalid_client_metadata", "invalid JSON body")
		return
	}
	if len(request.RedirectURIs) == 0 {
		sendError(c, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uris is required")
		return
	}

	// ClientName is a pointer (null when absent); only set it when the client
	// actually supplied a name.
	var clientName *string
	if request.ClientName != "" {
		clientName = &request.ClientName
	}
	clientID, err := uuid.NewV7()
	if err != nil {
		log.Error().Err(err).Msg("register failed: could not generate client_id")
		sendError(c, http.StatusInternalServerError, "server_error", "could not generate client_id")
		return
	}
	client := &store.Client{
		ClientID:     clientID,
		ClientName:   clientName,
		RedirectURIs: request.RedirectURIs,
		Type:         store.ClientTypeOAuth,
		CreatedAt:    time.Now(),
	}
	if err := h.appServices.Clients.SaveClient(c.Request.Context(), client); err != nil {
		log.Error().Err(err).Str("client_id", client.ClientID.String()).Msg("register failed: could not persist client")
		sendError(c, http.StatusInternalServerError, "server_error", "could not persist client")
		return
	}
	log.Info().Str("client_id", client.ClientID.String()).Int("redirect_uri_count", len(client.RedirectURIs)).Msg("client registered")

	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusCreated, gin.H{
		"client_id":                  client.ClientID,
		"client_name":                client.ClientName,
		"redirect_uris":              client.RedirectURIs,
		"token_endpoint_auth_method": "none",
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
	})
}

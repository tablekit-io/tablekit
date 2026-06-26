package oauth

import (
	"net/http"
	"net/url"
	"slices"
	"time"

	"core/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// authCodeTTL bounds how long an issued authorization code stays redeemable.
const authCodeTTL = 5 * time.Minute

// handleAuthorize implements GET /oauth/authorize.
//
// There is no login UI: authorization is auto-approved. Single-client pairing
// is enforced here — the first client to reach this endpoint successfully
// becomes THE paired client; any later, different client gets a plaintext
// "already paired" page instead of a code.
func (h *Handlers) handleAuthorize(c *gin.Context) {
	q := c.Request.URL.Query()
	responseType := q.Get("response_type")
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	challengeMethod := q.Get("code_challenge_method")
	scope := q.Get("scope")

	if responseType != "code" {
		authorizeError(c, "response_type must be \"code\"")
		return
	}
	if clientID == "" {
		authorizeError(c, "client_id is required")
		return
	}
	if redirectURI == "" {
		authorizeError(c, "redirect_uri is required")
		return
	}
	if codeChallenge == "" {
		authorizeError(c, "code_challenge is required")
		return
	}
	if challengeMethod != "S256" {
		authorizeError(c, "code_challenge_method must be S256")
		return
	}

	client, err := h.store.GetClient(clientID)
	if err != nil {
		authorizeError(c, "internal error loading client")
		return
	}
	if client == nil {
		authorizeError(c, "unknown client_id")
		return
	}
	if !slices.Contains(client.RedirectURIs, redirectURI) {
		authorizeError(c, "redirect_uri not registered for this client")
		return
	}

	// Pairing lock: succeeds only if unpaired or already this client.
	paired, err := h.store.Pair(clientID)
	if err != nil {
		authorizeError(c, "internal error during pairing")
		return
	}
	if !paired {
		c.String(http.StatusOK, "already paired")
		return
	}

	if scope == "" {
		scope = Scope
	}
	code := uuid.NewString()
	if err := h.store.PutCode(&store.AuthCode{
		Code:          code,
		ClientID:      clientID,
		RedirectURI:   redirectURI,
		CodeChallenge: codeChallenge,
		Scope:         scope,
		UserID:        UserID,
		ExpiresAt:     time.Now().Add(authCodeTTL),
	}); err != nil {
		authorizeError(c, "internal error issuing code")
		return
	}

	target, err := url.Parse(redirectURI)
	if err != nil {
		authorizeError(c, "malformed redirect_uri")
		return
	}
	rq := target.Query()
	rq.Set("code", code)
	if state != "" {
		rq.Set("state", state)
	}
	target.RawQuery = rq.Encode()
	c.Redirect(http.StatusFound, target.String())
}

// authorizeError renders a human-readable error (the user-agent is a browser
// here, not the client's HTTP stack).
func authorizeError(c *gin.Context, description string) {
	c.Header("Cache-Control", "no-store")
	c.String(http.StatusBadRequest, "OAuth error: %s", description)
}

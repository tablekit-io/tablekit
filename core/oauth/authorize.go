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
// There is no login UI: authorization is auto-approved. Pairing is enforced
// here via the store's pairing mode (once/indefinite/disabled): an
// already-paired client is always admitted, a new client is admitted only if
// the mode allows, otherwise it gets the "already paired" page that bounces
// back with an OAuth error. See `tablekit pairing` to change the mode.
func (h *Handlers) handleAuthorize(c *gin.Context) {
	// Hold the lock for the entire handler so pairing decisions cannot race.
	h.authorizeMu.Lock()
	defer h.authorizeMu.Unlock()

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

	client, err := h.appServices.Store.GetClient(clientID)
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

	// Pairing gate: allowed if already paired or the mode permits a new client.
	allowed, err := h.appServices.Store.TryPair(clientID)
	if err != nil {
		authorizeError(c, "internal error during pairing")
		return
	}
	if !allowed {
		renderAlreadyPaired(c, redirectURI, state)
		return
	}

	if scope == "" {
		scope = Scope
	}
	code := uuid.NewString()
	if err := h.appServices.Store.PutCode(&store.AuthCode{
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

// alreadyPairedRedirectDelay is how long the "already paired" page lingers
// before bouncing back to the client with an OAuth error.
const alreadyPairedRedirectDelay = 4

// renderAlreadyPaired shows a brief "already paired" page, then redirects back
// to the client's redirect_uri with error=access_denied (RFC 6749 §4.1.2.1) so
// the client surfaces a real failure instead of hanging. redirect_uri was
// already validated as registered for this client, so the bounce is safe.
func renderAlreadyPaired(c *gin.Context, redirectURI, state string) {
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Type", "text/html; charset=utf-8")

	target, err := url.Parse(redirectURI)
	if err != nil {
		c.String(http.StatusOK, "already paired")
		return
	}
	q := target.Query()
	q.Set("error", "access_denied")
	q.Set("error_description", "this server is already paired with another client")
	if state != "" {
		q.Set("state", state)
	}
	target.RawQuery = q.Encode()

	c.Status(http.StatusOK)
	if err := alreadyPairedTmpl.Execute(c.Writer, struct {
		Delay       int
		RedirectURL string
	}{
		Delay:       alreadyPairedRedirectDelay,
		RedirectURL: target.String(),
	}); err != nil {
		c.String(http.StatusInternalServerError, "already paired")
	}
}

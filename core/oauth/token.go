package oauth

import (
	"net/http"
	"time"

	"core/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleToken implements POST /oauth/token for the authorization_code and
// refresh_token grants. Client authentication is "none" (public client +
// PKCE), so the client is identified by the client_id form field.
func (h *Handlers) handleToken(c *gin.Context) {
	grantType := c.PostForm("grant_type")
	clientID := c.PostForm("client_id")

	if clientID == "" {
		sendError(c, http.StatusUnauthorized, "invalid_client", "client_id is required")
		return
	}
	client, err := h.appServices.Store.GetClient(clientID)
	if err != nil || client == nil {
		sendError(c, http.StatusUnauthorized, "invalid_client", "unknown client_id")
		return
	}

	switch grantType {
	case "authorization_code":
		h.authCodeGrant(c, client)
	case "refresh_token":
		h.refreshGrant(c, client)
	default:
		sendError(c, http.StatusBadRequest, "unsupported_grant_type",
			"only authorization_code and refresh_token are supported")
	}
}

// authCodeGrant redeems a one-time code (with PKCE) and opens a fresh refresh
// chain, returning an access+refresh pair.
func (h *Handlers) authCodeGrant(c *gin.Context, client *store.Client) {
	code := c.PostForm("code")
	codeVerifier := c.PostForm("code_verifier")
	redirectURI := c.PostForm("redirect_uri")

	if code == "" || codeVerifier == "" || redirectURI == "" {
		sendError(c, http.StatusBadRequest, "invalid_request",
			"code, code_verifier and redirect_uri are required")
		return
	}

	authCode, err := h.appServices.Store.ConsumeCode(code)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "server_error", "could not read code")
		return
	}
	if authCode == nil {
		sendError(c, http.StatusBadRequest, "invalid_grant", "unknown or used code")
		return
	}
	if time.Now().After(authCode.ExpiresAt) {
		sendError(c, http.StatusBadRequest, "invalid_grant", "code expired")
		return
	}
	if authCode.ClientID != client.ClientID {
		sendError(c, http.StatusBadRequest, "invalid_grant", "client mismatch")
		return
	}
	if authCode.RedirectURI != redirectURI {
		sendError(c, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
		return
	}
	if !verifyPKCE(codeVerifier, authCode.CodeChallenge) {
		sendError(c, http.StatusBadRequest, "invalid_grant", "PKCE check failed")
		return
	}

	chain := &store.Chain{
		ID:                uuid.NewString(),
		ClientID:          client.ClientID,
		UserID:            authCode.UserID,
		Scope:             authCode.Scope,
		RedirectURI:       authCode.RedirectURI,
		InvalidatedBefore: time.Unix(0, 0),
		CreatedAt:         time.Now(),
	}
	if err := h.appServices.Store.NewChain(chain); err != nil {
		sendError(c, http.StatusInternalServerError, "server_error", "could not open chain")
		return
	}

	h.issuePair(c, client.ClientID, chain.ID, authCode.Scope)
}

// refreshGrant rotates a refresh token. Reusing a superseded refresh token
// (iat <= the chain's invalidated_before cutoff) is treated as theft and
// revokes the entire chain (OAuth 2.1 BCP).
func (h *Handlers) refreshGrant(c *gin.Context, client *store.Client) {
	refreshToken := c.PostForm("refresh_token")
	if refreshToken == "" {
		sendError(c, http.StatusBadRequest, "invalid_request", "refresh_token required")
		return
	}

	claims, err := h.issuer.VerifyRefresh(refreshToken)
	if err != nil {
		sendError(c, http.StatusBadRequest, "invalid_grant", "invalid refresh token")
		return
	}
	if claims.CID != client.ClientID {
		sendError(c, http.StatusBadRequest, "invalid_grant", "token does not belong to client")
		return
	}

	chain, err := h.appServices.Store.GetChain(claims.Chain)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "server_error", "could not read chain")
		return
	}
	if chain == nil {
		sendError(c, http.StatusBadRequest, "invalid_grant", "unknown chain")
		return
	}
	if chain.Revoked {
		sendError(c, http.StatusBadRequest, "invalid_grant", "chain revoked")
		return
	}

	issuedAt := claims.IssuedAt.Time
	if !issuedAt.After(chain.InvalidatedBefore) {
		// Replay of an already-rotated token: kill the whole chain.
		_ = h.appServices.Store.RevokeChain(chain.ID)
		sendError(c, http.StatusBadRequest, "invalid_grant", "refresh token reuse detected")
		return
	}

	if err := h.appServices.Store.BumpCutoff(chain.ID, issuedAt); err != nil {
		sendError(c, http.StatusInternalServerError, "server_error", "could not rotate chain")
		return
	}

	h.issuePair(c, client.ClientID, chain.ID, chain.Scope)
}

// issuePair mints an access+refresh pair and writes the token response.
func (h *Handlers) issuePair(c *gin.Context, clientID, chainID, scope string) {
	access, err := h.issuer.IssueAccess(clientID, chainID, scope)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "server_error", "could not issue access token")
		return
	}
	refresh, _, err := h.issuer.IssueRefresh(clientID, chainID, scope)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "server_error", "could not issue refresh token")
		return
	}

	noStoreJSON(c, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"expires_in":    int(h.appServices.Config.AccessTTL.Seconds()),
		"scope":         scope,
	})
}

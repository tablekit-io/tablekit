package oauth

import (
	"net/http"
	"time"

	"core/services/oauth"
	"core/services/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// HandleToken implements POST /oauth/token for the authorization_code and
// refresh_token grants. Client authentication is "none" (public client +
// PKCE), so the client is identified by the client_id form field.
func (h *Handlers) HandleToken(c *gin.Context) {
	grantType := c.PostForm("grant_type")
	rawClientID := c.PostForm("client_id")

	if rawClientID == "" {
		sendError(c, http.StatusUnauthorized, "invalid_client", "client_id is required")
		return
	}
	clientID, err := uuid.Parse(rawClientID)
	if err != nil {
		log.Warn().Str("client_id", rawClientID).Msg("token rejected: unparseable client_id")
		sendError(c, http.StatusUnauthorized, "invalid_client", "unknown client_id")
		return
	}
	client, err := h.appServices.Clients.GetClient(c.Request.Context(), clientID)
	if err != nil {
		log.Error().Err(err).Str("client_id", clientID.String()).Msg("token: client lookup failed")
		sendError(c, http.StatusUnauthorized, "invalid_client", "unknown client_id")
		return
	}
	if client == nil {
		log.Warn().Str("client_id", clientID.String()).Msg("token rejected: unknown client_id")
		sendError(c, http.StatusUnauthorized, "invalid_client", "unknown client_id")
		return
	}

	switch grantType {
	case "authorization_code":
		h.authCodeGrant(c, client)
	case "refresh_token":
		h.refreshGrant(c, client)
	default:
		log.Warn().Str("client_id", clientID.String()).Str("grant_type", grantType).Msg("token rejected: unsupported grant_type")
		sendError(c, http.StatusBadRequest, "unsupported_grant_type",
			"only authorization_code and refresh_token are supported")
	}
}

// authCodeGrant redeems a one-time code (with PKCE) and opens a fresh refresh
// chain, returning an access+refresh pair.
func (h *Handlers) authCodeGrant(c *gin.Context, client *store.Client) {
	rawCode := c.PostForm("code")
	codeVerifier := c.PostForm("code_verifier")
	redirectURI := c.PostForm("redirect_uri")

	if rawCode == "" || codeVerifier == "" || redirectURI == "" {
		sendError(c, http.StatusBadRequest, "invalid_request",
			"code, code_verifier and redirect_uri are required")
		return
	}

	authCode, err := h.appServices.AuthCodes.ConsumeCode(c.Request.Context(), rawCode)
	if err != nil {
		log.Error().Err(err).Str("client_id", client.ClientID.String()).Msg("code grant: consume code failed")
		sendError(c, http.StatusInternalServerError, "server_error", "could not read code")
		return
	}
	if authCode == nil {
		log.Warn().Str("client_id", client.ClientID.String()).Msg("code grant rejected: unknown or used code")
		sendError(c, http.StatusBadRequest, "invalid_grant", "unknown or used code")
		return
	}
	if time.Now().After(authCode.ExpiresAt) {
		log.Warn().Str("client_id", client.ClientID.String()).Msg("code grant rejected: code expired")
		sendError(c, http.StatusBadRequest, "invalid_grant", "code expired")
		return
	}
	if authCode.ClientID != client.ClientID {
		log.Warn().Str("client_id", client.ClientID.String()).Str("code_client_id", authCode.ClientID.String()).Msg("code grant rejected: client mismatch")
		sendError(c, http.StatusBadRequest, "invalid_grant", "client mismatch")
		return
	}
	if authCode.RedirectURI != redirectURI {
		log.Warn().Str("client_id", client.ClientID.String()).Msg("code grant rejected: redirect_uri mismatch")
		sendError(c, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
		return
	}
	if !oauth.VerifyPKCE(codeVerifier, authCode.CodeChallenge) {
		log.Warn().Str("client_id", client.ClientID.String()).Msg("code grant rejected: PKCE check failed")
		sendError(c, http.StatusBadRequest, "invalid_grant", "PKCE check failed")
		return
	}

	chainID, err := uuid.NewV7()
	if err != nil {
		log.Error().Err(err).Str("client_id", client.ClientID.String()).Msg("code grant: chain id generation failed")
		sendError(c, http.StatusInternalServerError, "server_error", "could not open chain")
		return
	}
	chain := &store.Chain{
		ID:                chainID,
		ClientID:          client.ClientID,
		UserID:            authCode.UserID,
		Scope:             authCode.Scope,
		RedirectURI:       authCode.RedirectURI,
		InvalidatedBefore: time.Unix(0, 0),
		CreatedAt:         time.Now(),
	}
	if err := h.appServices.TokenChains.NewChain(c.Request.Context(), chain); err != nil {
		log.Error().Err(err).Str("client_id", client.ClientID.String()).Str("chain_id", chain.ID.String()).Msg("code grant: new chain failed")
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

	claims, err := h.appServices.Issuer.VerifyRefresh(refreshToken)
	if err != nil {
		log.Warn().Err(err).Str("client_id", client.ClientID.String()).Msg("refresh rejected: invalid refresh token")
		sendError(c, http.StatusBadRequest, "invalid_grant", "invalid refresh token")
		return
	}
	if claims.CID != client.ClientID.String() {
		log.Warn().Str("client_id", client.ClientID.String()).Msg("refresh rejected: token does not belong to client")
		sendError(c, http.StatusBadRequest, "invalid_grant", "token does not belong to client")
		return
	}

	chainID, err := uuid.Parse(claims.Chain)
	if err != nil {
		log.Warn().Str("client_id", client.ClientID.String()).Msg("refresh rejected: unparseable chain id")
		sendError(c, http.StatusBadRequest, "invalid_grant", "unknown chain")
		return
	}
	chain, err := h.appServices.TokenChains.GetChain(c.Request.Context(), chainID)
	if err != nil {
		log.Error().Err(err).Str("client_id", client.ClientID.String()).Str("chain_id", chainID.String()).Msg("refresh: get chain failed")
		sendError(c, http.StatusInternalServerError, "server_error", "could not read chain")
		return
	}
	if chain == nil {
		log.Warn().Str("client_id", client.ClientID.String()).Str("chain_id", chainID.String()).Msg("refresh rejected: unknown chain")
		sendError(c, http.StatusBadRequest, "invalid_grant", "unknown chain")
		return
	}
	if chain.Revoked() {
		log.Warn().Str("client_id", client.ClientID.String()).Str("chain_id", chainID.String()).Msg("refresh rejected: chain revoked")
		sendError(c, http.StatusBadRequest, "invalid_grant", "chain revoked")
		return
	}

	issuedAt := claims.IssuedAt.Time
	if !issuedAt.After(chain.InvalidatedBefore) {
		// Replay of an already-rotated token: kill the whole chain.
		log.Warn().Str("client_id", client.ClientID.String()).Str("chain_id", chain.ID.String()).Msg("refresh token reuse detected, revoking chain")
		if revokeErr := h.appServices.TokenChains.RevokeChain(c.Request.Context(), chain.ID); revokeErr != nil {
			log.Error().Err(revokeErr).Str("chain_id", chain.ID.String()).Msg("refresh: revoke chain failed during reuse handling")
		}
		sendError(c, http.StatusBadRequest, "invalid_grant", "refresh token reuse detected")
		return
	}

	if err := h.appServices.TokenChains.BumpCutoff(c.Request.Context(), chain.ID, issuedAt); err != nil {
		log.Error().Err(err).Str("chain_id", chain.ID.String()).Msg("refresh: bump cutoff failed")
		sendError(c, http.StatusInternalServerError, "server_error", "could not rotate chain")
		return
	}

	h.issuePair(c, client.ClientID, chain.ID, chain.Scope)
}

// issuePair mints an access+refresh pair and writes the token response.
func (h *Handlers) issuePair(c *gin.Context, clientID, chainID uuid.UUID, scope string) {
	access, err := h.appServices.Issuer.IssueAccess(clientID.String(), chainID.String(), scope)
	if err != nil {
		log.Error().Err(err).Str("client_id", clientID.String()).Str("chain_id", chainID.String()).Msg("issue: access token minting failed")
		sendError(c, http.StatusInternalServerError, "server_error", "could not issue access token")
		return
	}
	refresh, _, err := h.appServices.Issuer.IssueRefresh(clientID.String(), chainID.String(), scope)
	if err != nil {
		log.Error().Err(err).Str("client_id", clientID.String()).Str("chain_id", chainID.String()).Msg("issue: refresh token minting failed")
		sendError(c, http.StatusInternalServerError, "server_error", "could not issue refresh token")
		return
	}
	log.Info().Str("client_id", clientID.String()).Str("chain_id", chainID.String()).Str("scope", scope).Msg("tokens issued")

	noStoreJSON(c, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"expires_in":    int(h.appServices.Config.AccessTTL.Seconds()),
		"scope":         scope,
	})
}

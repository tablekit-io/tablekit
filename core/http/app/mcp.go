package app

import (
	"context"
	"net/http"
	"strings"

	"core/mcp"
	"core/services"
	"core/services/oauth"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/auth"
)

// MCPRoute wraps the MCP Streamable HTTP handler with the OAuth bearer-token
// middleware. The verifier validates our access JWT; the middleware enforces
// expiration + scope and binds the session to the token's user, and emits a
// WWW-Authenticate header pointing at the protected-resource metadata on 401.
func MCPRoute(appServices *services.Services) http.Handler {
	issuer := appServices.Issuer
	verifier := func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		// A CLI-minted static token is marked by TokenPrefix: validate it as a
		// static JWT and reject it if its store row is missing or revoked.
		if rawJWT, ok := strings.CutPrefix(token, oauth.TokenPrefix); ok {
			claims, err := issuer.VerifyStatic(rawJWT)
			if err != nil {
				return nil, auth.ErrInvalidToken
			}
			tokenID, err := uuid.Parse(claims.ID)
			if err != nil {
				return nil, auth.ErrInvalidToken
			}
			staticToken, err := appServices.StaticTokens.GetStaticToken(ctx, tokenID)
			if err != nil || staticToken == nil || staticToken.Revoked() {
				return nil, auth.ErrInvalidToken
			}
			return &auth.TokenInfo{
				UserID:     oauth.UserID,
				Scopes:     []string{claims.Scope},
				Expiration: claims.ExpiresAt.Time,
				// Carry the client id (claims.cid) so the MCP request audit log can
				// attribute each request to the client its token was issued to.
				Extra: map[string]any{"client_id": claims.CID},
			}, nil
		}

		claims, err := issuer.VerifyAccess(token)
		if err != nil {
			return nil, auth.ErrInvalidToken
		}
		return &auth.TokenInfo{
			UserID:     oauth.UserID,
			Scopes:     []string{claims.Scope},
			Expiration: claims.ExpiresAt.Time,
			Extra:      map[string]any{"client_id": claims.CID},
		}, nil
	}

	return auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{
		Scopes:              []string{oauth.Scope},
		ResourceMetadataURL: appServices.Config.PublicBaseURL + "/.well-known/oauth-protected-resource",
	})(mcp.StreamableHandler(appServices))
}

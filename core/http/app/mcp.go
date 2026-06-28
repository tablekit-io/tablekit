package app

import (
	"context"
	"net/http"
	"strings"

	"core/mcp"
	"core/services"
	"core/services/oauth"

	"github.com/modelcontextprotocol/go-sdk/auth"
)

// MCPRoute wraps the MCP Streamable HTTP handler with the OAuth bearer-token
// middleware. The verifier validates our access JWT; the middleware enforces
// expiration + scope and binds the session to the token's user, and emits a
// WWW-Authenticate header pointing at the protected-resource metadata on 401.
func MCPRoute(appServices *services.Services) http.Handler {
	issuer := appServices.Issuer
	verifier := func(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		// A CLI-minted bearer token is marked by TokenPrefix: validate it as a
		// bearer JWT and reject it if its store row is missing or revoked.
		if rawJWT, ok := strings.CutPrefix(token, oauth.TokenPrefix); ok {
			claims, err := issuer.VerifyBearer(rawJWT)
			if err != nil {
				return nil, auth.ErrInvalidToken
			}
			bearerToken, err := appServices.Store.GetBearerToken(claims.ID)
			if err != nil || bearerToken == nil || bearerToken.Revoked {
				return nil, auth.ErrInvalidToken
			}
			return &auth.TokenInfo{
				UserID:     oauth.UserID,
				Scopes:     []string{claims.Scope},
				Expiration: claims.ExpiresAt.Time,
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
		}, nil
	}

	return auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{
		Scopes:              []string{oauth.Scope},
		ResourceMetadataURL: appServices.Config.PublicBaseURL + "/.well-known/oauth-protected-resource",
	})(mcp.StreamableHandler(appServices.Engine))
}

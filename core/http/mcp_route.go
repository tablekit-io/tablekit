package http

import (
	"context"
	nethttp "net/http"

	"core/http/app/oauth"
	"core/mcp"
	"core/services"

	"github.com/modelcontextprotocol/go-sdk/auth"
)

// mcpRoute wraps the MCP Streamable HTTP handler with the OAuth bearer-token
// middleware. The verifier validates our access JWT; the middleware enforces
// expiration + scope and binds the session to the token's user, and emits a
// WWW-Authenticate header pointing at the protected-resource metadata on 401.
func mcpRoute(appServices *services.Services, issuer *oauth.Issuer) nethttp.Handler {
	verifier := func(_ context.Context, token string, _ *nethttp.Request) (*auth.TokenInfo, error) {
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
	})(mcp.StreamableHandler())
}

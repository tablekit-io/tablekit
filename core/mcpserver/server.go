// Package mcpserver exposes the MCP server over Streamable HTTP, guarded by the
// OAuth bearer-token middleware. It registers a single hello_world tool.
package mcpserver

import (
	"context"
	"net/http"

	"core/config"
	"core/oauth"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// helloInput is the hello_world tool's argument schema. Name is optional.
type helloInput struct {
	Name string `json:"name,omitempty" jsonschema:"name to greet; defaults to world"`
}

// helloOutput is the hello_world tool's structured result. Because Out is a
// struct (not any), the SDK generates the tool's outputSchema from it and
// populates CallToolResult.StructuredContent with this value automatically.
type helloOutput struct {
	Greeting string `json:"greeting" jsonschema:"the greeting message"`
}

// helloWorld returns a greeting both as human-readable text and as structured
// output validated against helloOutput's schema.
func helloWorld(_ context.Context, _ *mcp.CallToolRequest, in helloInput) (*mcp.CallToolResult, helloOutput, error) {
	name := in.Name
	if name == "" {
		name = "world"
	}
	greeting := "Hello, " + name + "!"
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: greeting}},
	}, helloOutput{Greeting: greeting}, nil
}

// newServer builds the MCP server and registers tools.
func newServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "tablekit",
		Title:   "tablekit MCP server",
		Version: "0.1.0",
	}, nil)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "hello_world",
		Description: "Returns a friendly greeting, optionally addressed to a name.",
	}, helloWorld)
	return s
}

// Handler returns the bearer-protected Streamable HTTP handler for /mcp.
// The verifier validates our OAuth access JWT; the middleware enforces
// expiration + scope and binds the session to the token's user, and emits a
// WWW-Authenticate header pointing at the protected-resource metadata on 401.
func Handler(cfg *config.Config, issuer *oauth.Issuer) http.Handler {
	server := newServer()

	streamable := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{
			Stateless: true,
			// This server is deployed behind a reverse proxy and reached via a
			// public hostname (PUBLIC_BASE_URL). The SDK's DNS-rebinding guard
			// rejects requests whose Host header is non-loopback when the proxy
			// forwards over 127.0.0.1, which 403s every /mcp call (e.g. ChatGPT
			// gets "there was an issue" and no tools load). That guard targets
			// browser attacks on local dev servers; /mcp already requires a
			// bearer token, so disable it here.
			DisableLocalhostProtection: true,
		},
	)

	verifier := func(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
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
		ResourceMetadataURL: cfg.PublicBaseURL + "/.well-known/oauth-protected-resource",
	})(streamable)
}

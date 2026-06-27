// Package mcpserver exposes the MCP server over Streamable HTTP, guarded by the
// OAuth bearer-token middleware. It registers the hello_world tool, the
// hello_world_interactive donut widget (an MCP Apps UI resource) and its
// app-only hello_world_interactive_data loader.
package mcp

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"core/http/app/mcp/widgets"
	"core/http/app/oauth"
	"core/services"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// pointer returns a pointer to v, for the optional *bool annotation hints.
func pointer[T any](v T) *T { return &v }

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

// helloInteractiveName is the widget template name (the @tablekit/widgets
// manifest key) and the tool name stem the donut demo is built around.
const helloInteractiveWidget = "hello_world_interactive"

// helloInteractiveInput is the hello_world_interactive tool's argument schema.
type helloInteractiveInput struct{}

// helloInteractiveOutput is a thin discriminator: the widget shares no state
// with the agent, so the tool just names itself. The host renders the linked
// ui:// widget, which fetches its own data over the MCP Apps bridge.
type helloInteractiveOutput struct {
	Tool string `json:"tool" jsonschema:"the tool that produced this result"`
}

// helloInteractive renders the interactive donut widget. The structured result
// is only a discriminator; the real payload is loaded by the widget calling the
// app-only hello_world_interactive_data tool over the bridge.
func helloInteractive(_ context.Context, _ *mcp.CallToolRequest, _ helloInteractiveInput) (*mcp.CallToolResult, helloInteractiveOutput, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: "Rendering an interactive donut chart with random example data.",
		}},
	}, helloInteractiveOutput{Tool: helloInteractiveWidget}, nil
}

// dataInput is the hello_world_interactive_data loader's argument schema.
type dataInput struct {
	Slices int `json:"slices,omitempty" jsonschema:"number of donut slices (1-8); defaults to 5"`
}

// dataSlice is one donut slice. Lowercase json tags match what the widget reads
// off structuredContent.
type dataSlice struct {
	Label string  `json:"label" jsonschema:"the slice label"`
	Value float64 `json:"value" jsonschema:"the slice value"`
}

// dataOutput is the loader's structured result: the random slices to plot.
type dataOutput struct {
	Data []dataSlice `json:"data" jsonschema:"the donut slices"`
}

// helloInteractiveData is the example data loader: it returns a random dataset
// for the donut. App-only — hidden from the model, called only by the widget
// over the MCP Apps bridge.
func helloInteractiveData(_ context.Context, _ *mcp.CallToolRequest, in dataInput) (*mcp.CallToolResult, dataOutput, error) {
	n := in.Slices
	if n < 1 {
		n = 5
	}
	if n > 8 {
		n = 8
	}
	slices := make([]dataSlice, n)
	for i := 0; i < n; i++ {
		slices[i] = dataSlice{
			Label: fmt.Sprintf("Category %c", 'A'+i),
			Value: float64(rand.Intn(91) + 10), // 10..100
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Generated %d random donut slices.", n),
		}},
	}, dataOutput{Data: slices}, nil
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
		// Without these hints the MCP spec defaults are conservative
		// (destructive + open-world + write), which clients like ChatGPT
		// surface as scary badges. This tool only computes a string.
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			DestructiveHint: pointer(false),
			OpenWorldHint:   pointer(false),
		},
	}, helloWorld)

	// Register the built widget templates as ui:// resources the host can render
	// in a sandboxed iframe. Empty until @tablekit/widgets is built.
	for _, r := range widgets.Resources() {
		uri := r.URI
		mime := r.MIMEType
		html := r.HTML
		s.AddResource(
			&mcp.Resource{Name: r.Name, URI: uri, MIMEType: mime},
			func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{URI: uri, MIMEType: mime, Text: html},
					},
				}, nil
			},
		)
	}

	// hello_world_interactive: model-facing, renders the donut widget. The
	// _meta.ui.resourceUri links the widget the host prefetches and renders; it's
	// resolved from the build manifest (empty until widgets are built, in which
	// case the tool simply advertises no widget).
	interactive := &mcp.Tool{
		Name:        "hello_world_interactive",
		Description: "Renders an interactive donut chart with random example data, using MCP Apps.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			DestructiveHint: pointer(false),
			OpenWorldHint:   pointer(false),
		},
	}
	if uri := widgets.WidgetURI(helloInteractiveWidget); uri != "" {
		interactive.Meta = mcp.Meta{"ui": map[string]any{"resourceUri": uri}}
	}
	mcp.AddTool(s, interactive, helloInteractive)

	// hello_world_interactive_data: the example data loader. _meta.ui.visibility
	// = ['app'] marks it app-only — the host hides it from the model and only
	// honours it when the widget calls it over the bridge.
	dataTool := &mcp.Tool{
		Name:        "hello_world_interactive_data",
		Description: "Returns random example data for the hello_world_interactive donut. App-only: called by the widget over the MCP Apps bridge, hidden from the agent.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: pointer(false),
			OpenWorldHint:   pointer(false),
		},
	}
	dataTool.Meta = mcp.Meta{"ui": map[string]any{"visibility": []string{"app"}}}
	mcp.AddTool(s, dataTool, helloInteractiveData)

	return s
}

// Handler returns the bearer-protected Streamable HTTP handler for /mcp.
// The verifier validates our OAuth access JWT; the middleware enforces
// expiration + scope and binds the session to the token's user, and emits a
// WWW-Authenticate header pointing at the protected-resource metadata on 401.
func Handler(appServices *services.Services, issuer *oauth.Issuer) http.Handler {
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
		ResourceMetadataURL: appServices.Config.PublicBaseURL + "/.well-known/oauth-protected-resource",
	})(streamable)
}

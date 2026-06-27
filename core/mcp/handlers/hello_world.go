package handlers

import (
	"context"

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

// registerHelloWorld adds the hello_world tool to the server.
func registerHelloWorld(s *mcp.Server) {
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
}

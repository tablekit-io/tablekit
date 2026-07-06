// Package getexporturl implements the app-only get_export_url MCP tool.
package getexporturl

import (
	"context"
	_ "embed"
	"fmt"

	"core/helpers"
	"core/mcp/handlers/shared"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed schema.json
var schemaJSON []byte

// input is the get_export_url tool's argument schema.
type input struct {
	QueryKey string `json:"query_key"`
	Format   string `json:"format"`
}

// output carries the short-lived signed download URL.
type output struct {
	URL string `json:"url" jsonschema:"a short-lived signed URL that downloads the full result in the requested format"`
}

// Register adds the get_export_url tool. It only signs a URL (no DB access of its
// own), so it is read-only and not open-world. App-only: the
// _meta.ui.visibility=['app'] hides it from the model — the chart widget's
// download button calls it over the MCP Apps bridge and opens the link in the
// user's real browser. An export link is a user-facing download affordance, not
// something the agent needs to reason about.
func Register(s *mcp.Server, deps shared.Deps) {
	tool := &mcp.Tool{
		Name:        "get_export_url",
		Description: "Returns a short-lived signed URL that downloads the full result of a stored query as CSV or JSON. App-only: called by the chart widget's download button over the MCP Apps bridge, hidden from the agent.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}
	tool.Meta = mcp.Meta{"ui": map[string]any{"visibility": []string{"app"}}}
	tool.InputSchema = shared.InputSchema[input](schemaJSON)
	mcp.AddTool(s, tool, handle(deps))
}

// handle returns a short-lived signed URL that downloads the full result of a
// stored query as CSV or JSON. The URL points at the /exports endpoint, which
// re-runs the stored query when fetched; the link is the only credential needed.
func handle(deps shared.Deps) mcp.ToolHandlerFor[input, output] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in input) (*mcp.CallToolResult, output, error) {
		if in.Format != "csv" && in.Format != "json" {
			return nil, output{}, fmt.Errorf("format must be \"csv\" or \"json\", got %q", in.Format)
		}

		descriptor, err := deps.Queries.Get(ctx, in.QueryKey)
		if err != nil {
			return nil, output{}, err
		}
		if descriptor == nil {
			return nil, output{}, fmt.Errorf("unknown query_key %q (run query_database first)", in.QueryKey)
		}

		token, err := deps.Issuer.IssueExport(in.QueryKey)
		if err != nil {
			return nil, output{}, err
		}
		url := fmt.Sprintf("%s/exports/%s/%s", deps.PublicBaseURL, in.Format, token)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Export link (%s, valid ~5 minutes): %s", in.Format, url),
			}},
		}, output{URL: url}, nil
	}
}

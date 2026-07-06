package handlers

import (
	"context"
	"fmt"

	"core/helpers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// getExportURLInput is the get_export_url tool's argument schema.
type getExportURLInput struct {
	QueryKey string `json:"query_key" jsonschema:"the result_key returned by query_database"`
	Format   string `json:"format" jsonschema:"the export format: \"csv\" or \"json\""`
}

// getExportURLOutput carries the short-lived signed download URL.
type getExportURLOutput struct {
	URL string `json:"url" jsonschema:"a short-lived signed URL that downloads the full result in the requested format"`
}

// getExportURL returns a short-lived signed URL that downloads the full result of
// a stored query as CSV or JSON. The URL points at the /exports endpoint, which
// re-runs the stored query when fetched; the link is the only credential needed.
func (h *Handlers) getExportURL(ctx context.Context, _ *mcp.CallToolRequest, in getExportURLInput) (*mcp.CallToolResult, getExportURLOutput, error) {
	if in.Format != "csv" && in.Format != "json" {
		return nil, getExportURLOutput{}, fmt.Errorf("format must be \"csv\" or \"json\", got %q", in.Format)
	}

	descriptor, err := h.Queries.Get(ctx, in.QueryKey)
	if err != nil {
		return nil, getExportURLOutput{}, err
	}
	if descriptor == nil {
		return nil, getExportURLOutput{}, fmt.Errorf("unknown query_key %q (run query_database first)", in.QueryKey)
	}

	token, err := h.Issuer.IssueExport(in.QueryKey)
	if err != nil {
		return nil, getExportURLOutput{}, err
	}
	url := fmt.Sprintf("%s/exports/%s/%s", h.PublicBaseURL, in.Format, token)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Export link (%s, valid ~5 minutes): %s", in.Format, url),
		}},
	}, getExportURLOutput{URL: url}, nil
}

// registerGetExportURL adds the get_export_url tool. It only signs a URL (no DB
// access of its own), so it is read-only and not open-world. App-only: the
// _meta.ui.visibility=['app'] hides it from the model — the chart widget's
// download button calls it over the MCP Apps bridge and opens the link in the
// user's real browser. An export link is a user-facing download affordance, not
// something the agent needs to reason about.
func (h *Handlers) registerGetExportURL(s *mcp.Server) {
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
	mcp.AddTool(s, tool, h.getExportURL)
}

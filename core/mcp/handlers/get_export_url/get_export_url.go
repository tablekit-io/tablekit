// Package getexporturl implements the app-only get_export_url MCP tool.
package getexporturl

import (
	"context"
	_ "embed"
	"fmt"

	"core/helpers"
	"core/mcp/handlers/shared"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
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
	tool.Meta = shared.WidgetBridgeMeta()
	tool.InputSchema = shared.InputSchema[input](schemaJSON)
	mcp.AddTool(s, tool, handle(deps))
}

// handle returns a short-lived signed URL that downloads the full result of a
// stored query as CSV or JSON. The URL points at the /exports endpoint, which
// re-runs the stored query when fetched; the link is the only credential needed.
func handle(deps shared.Deps) mcp.ToolHandlerFor[input, output] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in input) (*mcp.CallToolResult, output, error) {
		if in.Format != "csv" && in.Format != "json" {
			log.Warn().Str("format", in.Format).Msg("get_export_url invalid format")
			return nil, output{}, fmt.Errorf("format must be \"csv\" or \"json\", got %q", in.Format)
		}

		queryID, err := uuid.Parse(in.QueryKey)
		if err != nil {
			log.Warn().Str("query_key", in.QueryKey).Msg("get_export_url unknown query_key")
			return nil, output{}, fmt.Errorf("unknown query_key %q (run query_database first)", in.QueryKey)
		}
		descriptor, err := deps.Queries.Get(ctx, queryID)
		if err != nil {
			log.Error().Str("query_key", in.QueryKey).Err(err).Msg("get_export_url descriptor load failed")
			return nil, output{}, err
		}
		if descriptor == nil {
			log.Warn().Str("query_key", in.QueryKey).Msg("get_export_url unknown query_key")
			return nil, output{}, fmt.Errorf("unknown query_key %q (run query_database first)", in.QueryKey)
		}

		// Fail at link-mint time if the database was repointed, rather than handing
		// out a link that only errors when downloaded. The exports endpoint verifies
		// again when the link is fetched.
		if _, err := deps.Databases.Verify(ctx, descriptor.DatabaseID); err != nil {
			log.Warn().Str("query_key", in.QueryKey).Err(err).Msg("get_export_url verify failed (repointed?)")
			return nil, output{}, err
		}

		token, err := deps.Issuer.IssueExport(in.QueryKey)
		if err != nil {
			log.Error().Str("query_key", in.QueryKey).Err(err).Msg("get_export_url token issue failed")
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

package handlers

import (
	"context"
	"fmt"

	"core/helpers"
	"core/mcp/ui"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// proportionalLayer is one grouping ring: each distinct value of the column
// becomes a slice, with values summed per group. Layers go innermost first.
type proportionalLayer struct {
	DiscriminatorProp string `json:"discriminator_prop" jsonschema:"the column whose distinct values become slices (values are summed per group)"`
}

// renderProportionalInput is the render_proportional_chart tool's argument schema.
type renderProportionalInput struct {
	QueryKey    string              `json:"query_key" jsonschema:"the result_key returned by run_query"`
	Display     string              `json:"display,omitempty" jsonschema:"pie or donut (default donut)"`
	ValueProp   string              `json:"value_prop" jsonschema:"the column holding the numeric value of each slice"`
	ValuePrefix string              `json:"value_prefix,omitempty" jsonschema:"text shown before each value, e.g. \"$\""`
	ValueSuffix string              `json:"value_suffix,omitempty" jsonschema:"text shown after each value, e.g. \"%\" or \" users\""`
	Layers      []proportionalLayer `json:"layers" jsonschema:"one or more grouping layers, innermost ring first (at least one)"`
}

// renderProportionalChart renders the stored query as a proportional (pie/donut)
// chart. The structured result is only a discriminator; the linked widget reads
// this tool's arguments and loads rows via the app-only fetch_chart_data.
func (h *Handlers) renderProportionalChart(ctx context.Context, _ *mcp.CallToolRequest, in renderProportionalInput) (*mcp.CallToolResult, chartRenderOutput, error) {
	if in.ValueProp == "" {
		return nil, chartRenderOutput{}, fmt.Errorf("value_prop is required")
	}
	if len(in.Layers) == 0 {
		return nil, chartRenderOutput{}, fmt.Errorf("at least one layer is required")
	}
	if err := h.requireQuery(ctx, in.QueryKey); err != nil {
		return nil, chartRenderOutput{}, err
	}
	return chartRenderResult("render_proportional_chart", "proportional chart")
}

// registerRenderProportionalChart adds the proportional chart tool, linking the
// shared chart widget.
func (h *Handlers) registerRenderProportionalChart(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "render_proportional_chart",
		Description: "Renders a stored query as a proportional chart (pie/donut). Pass the result_key from run_query, the value column, and one or more grouping layers (innermost ring first). The chart widget loads the rows itself via the app-only fetch_chart_data.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}
	if uri := ui.WidgetURI(chartWidget); uri != "" {
		tool.Meta = mcp.Meta{"ui": map[string]any{"resourceUri": uri}}
	}
	mcp.AddTool(s, tool, h.renderProportionalChart)
}

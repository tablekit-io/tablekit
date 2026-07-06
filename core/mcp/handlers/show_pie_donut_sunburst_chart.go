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

// showPieDonutSunburstChartInput is the show_pie_donut_sunburst_chart tool's argument schema.
type showPieDonutSunburstChartInput struct {
	QueryKey    string              `json:"query_key" jsonschema:"the result_key returned by query_database"`
	Display     string              `json:"display,omitempty" jsonschema:"pie or donut (default donut)"`
	ValueProp   string              `json:"value_prop" jsonschema:"the column holding the numeric value of each slice"`
	ValuePrefix string              `json:"value_prefix,omitempty" jsonschema:"text shown before each value, e.g. \"$\""`
	ValueSuffix string              `json:"value_suffix,omitempty" jsonschema:"text shown after each value, e.g. \"%\" or \" users\""`
	Layers      []proportionalLayer `json:"layers" jsonschema:"one or more grouping layers, innermost ring first (at least one)"`
}

// showPieDonutSunburstChart renders the stored query as a proportional (pie/donut/
// sunburst) chart. The structured result is only a discriminator; the linked widget
// reads this tool's arguments and loads rows via the app-only fetch_chart_data.
func (h *Handlers) showPieDonutSunburstChart(ctx context.Context, _ *mcp.CallToolRequest, in showPieDonutSunburstChartInput) (*mcp.CallToolResult, chartRenderOutput, error) {
	if in.ValueProp == "" {
		return nil, chartRenderOutput{}, fmt.Errorf("value_prop is required")
	}
	if len(in.Layers) == 0 {
		return nil, chartRenderOutput{}, fmt.Errorf("at least one layer is required")
	}
	if err := h.requireQuery(ctx, in.QueryKey); err != nil {
		return nil, chartRenderOutput{}, err
	}
	return chartRenderResult("show_pie_donut_sunburst_chart", "pie/donut/sunburst chart")
}

// registerShowPieDonutSunburstChart adds the proportional chart tool, linking the
// shared chart widget.
func (h *Handlers) registerShowPieDonutSunburstChart(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "show_pie_donut_sunburst_chart",
		Description: "Use this for pie or donut charts. Shows a proportional chart visualization widget for a result_key received from query_database. Both chart types support stacking, stacking will result in a sunburst chart. Needs the value column and one or more grouping layers (inner-most ring first). Pass the result_key from query_database along with the columns & grouping. The chart widget loads the rows itself using the result_key. Note: users can view original SQL in the rendered chart widget, also the table of data which they can download as JSON or CSV.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}
	if uri := ui.WidgetURI(chartWidget); uri != "" {
		tool.Meta = mcp.Meta{"ui": map[string]any{"resourceUri": uri}}
	}
	mcp.AddTool(s, tool, h.showPieDonutSunburstChart)
}

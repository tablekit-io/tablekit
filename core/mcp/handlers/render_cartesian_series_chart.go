package handlers

import (
	"context"
	"fmt"

	"core/helpers"
	"core/mcp/ui"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// chartWidget is the @tablekit/widgets template both render tools link. The
// widget reads the render tool's arguments (the axis/series mapping) from the
// host, discriminates on the tool name in the result, and calls fetch_chart_data
// over the bridge to pull the rows it renders.
const chartWidget = "chart_renderer"

// cartesianAxis maps a result column to the X (category) axis.
type cartesianAxis struct {
	Prop      string `json:"prop" jsonschema:"the result column for the X/category axis"`
	AxesLabel string `json:"axes_label" jsonschema:"the X axis label"`
}

// cartesianSeries maps a result column to one Y series with its display options.
type cartesianSeries struct {
	Prop       string  `json:"prop" jsonschema:"the result column for this Y/value series"`
	AxesLabel  string  `json:"axes_label" jsonschema:"the legend + axis label for this series"`
	DisplayAs  string  `json:"display_as,omitempty" jsonschema:"how to draw this series: line, area or bar (default bar)"`
	Shape      string  `json:"shape,omitempty" jsonschema:"line shape: line, discrete or curve (default discrete; ignored for bars)"`
	ColorHue   float64 `json:"color_hue,omitempty" jsonschema:"hue 0-360; omit for a stable auto hue derived from axes_label"`
	StackGroup string  `json:"stack_group,omitempty" jsonschema:"series sharing a stack_group are stacked; others sit side by side"`
}

// renderCartesianInput is the render_cartesian_series_chart tool's argument schema.
type renderCartesianInput struct {
	QueryKey string            `json:"query_key" jsonschema:"the result_key returned by run_query"`
	FlipAxes bool              `json:"flip_axes,omitempty" jsonschema:"draw the chart horizontally when true"`
	X        cartesianAxis     `json:"x" jsonschema:"the X/category axis mapping"`
	Y        []cartesianSeries `json:"y" jsonschema:"one or more Y/value series to plot (at least one)"`
}

// renderCartesianSeriesChart renders the stored query as a cartesian (bar/line/
// area) chart. Like the donut demo the structured result is only a discriminator:
// the linked widget reads this tool's arguments and loads its own data via the
// app-only fetch_chart_data.
func (h *Handlers) renderCartesianSeriesChart(ctx context.Context, _ *mcp.CallToolRequest, in renderCartesianInput) (*mcp.CallToolResult, chartRenderOutput, error) {
	if len(in.Y) == 0 {
		return nil, chartRenderOutput{}, fmt.Errorf("at least one y series is required")
	}
	if err := h.requireQuery(ctx, in.QueryKey); err != nil {
		return nil, chartRenderOutput{}, err
	}
	return chartRenderResult("render_cartesian_series_chart", "cartesian series chart")
}

// registerRenderCartesianSeriesChart adds the cartesian chart tool, linking the
// shared chart widget.
func (h *Handlers) registerRenderCartesianSeriesChart(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "render_cartesian_series_chart",
		Description: "Renders a stored query as a cartesian chart (bar/line/area) with one X axis and one or more Y series. Pass the result_key from run_query plus the axis/series mapping. The chart widget loads the rows itself via the app-only fetch_chart_data.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}
	if uri := ui.WidgetURI(chartWidget); uri != "" {
		tool.Meta = mcp.Meta{"ui": map[string]any{"resourceUri": uri}}
	}
	mcp.AddTool(s, tool, h.renderCartesianSeriesChart)
}

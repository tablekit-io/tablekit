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

// showBarLineAreaChartInput is the show_bar_line_area_chart tool's argument schema.
type showBarLineAreaChartInput struct {
	QueryKey string            `json:"query_key" jsonschema:"the result_key returned by query_database"`
	FlipAxes bool              `json:"flip_axes,omitempty" jsonschema:"draw the chart horizontally when true"`
	X        cartesianAxis     `json:"x" jsonschema:"the X/category axis mapping"`
	Y        []cartesianSeries `json:"y" jsonschema:"one or more Y/value series to plot (at least one)"`
}

// showBarLineAreaChart renders the stored query as a bar/line/area chart. Like the
// donut demo the structured result is only a discriminator: the linked widget reads
// this tool's arguments and loads its own data via the app-only fetch_chart_data.
func (h *Handlers) showBarLineAreaChart(ctx context.Context, _ *mcp.CallToolRequest, in showBarLineAreaChartInput) (*mcp.CallToolResult, chartRenderOutput, error) {
	if len(in.Y) == 0 {
		return nil, chartRenderOutput{}, fmt.Errorf("at least one y series is required")
	}
	if err := h.requireQuery(ctx, in.QueryKey); err != nil {
		return nil, chartRenderOutput{}, err
	}
	return chartRenderResult("show_bar_line_area_chart", "bar/line/area chart")
}

// registerShowBarLineAreaChart adds the bar/line/area chart tool, linking the
// shared chart widget.
func (h *Handlers) registerShowBarLineAreaChart(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "show_bar_line_area_chart",
		Description: "Use this for bar charts, line charts, area charts or any combination of them. Shows a chart visualization widget for a result_key received from query_database. All chart types support stacking. Needs one X axis column and one or more Y series. Pass the result_key from query_database plus the axis/series mapping. The chart widget loads the rows itself using the result_key. Note: users can view original SQL in the rendered chart widget, also the table of data which they can download as JSON or CSV.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}
	if uri := ui.WidgetURI(chartWidget); uri != "" {
		tool.Meta = mcp.Meta{"ui": map[string]any{"resourceUri": uri}}
	}
	mcp.AddTool(s, tool, h.showBarLineAreaChart)
}

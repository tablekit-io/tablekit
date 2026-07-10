// Package showbarlineareachart implements the show_bar_line_area_chart MCP tool.
package showbarlineareachart

import (
	"context"
	_ "embed"
	"fmt"

	"core/helpers"
	"core/mcp/handlers/shared"
	"core/mcp/ui"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

//go:embed schema.json
var schemaJSON []byte

// cartesianAxis maps a result column to the X (category) axis.
type cartesianAxis struct {
	Prop      string `json:"prop"`
	AxesLabel string `json:"axes_label"`
}

// cartesianSeries maps a result column to one Y series with its display options.
type cartesianSeries struct {
	Prop        string  `json:"prop"`
	AxesLabel   string  `json:"axes_label"`
	ValuePrefix string  `json:"value_prefix,omitempty"`
	ValueSuffix string  `json:"value_suffix,omitempty"`
	DisplayAs   string  `json:"display_as,omitempty"`
	Shape       string  `json:"shape,omitempty"`
	ColorHue    float64 `json:"color_hue,omitempty"`
	StackGroup  string  `json:"stack_group,omitempty"`
}

// input is the show_bar_line_area_chart tool's argument schema. Descriptions and
// the display_as/shape enums live in schema.json; the struct only decodes.
type input struct {
	QueryKey    string            `json:"query_key"`
	FlipAxes    bool              `json:"flip_axes,omitempty"`
	ValuePrefix string            `json:"value_prefix,omitempty"`
	ValueSuffix string            `json:"value_suffix,omitempty"`
	X           cartesianAxis     `json:"x"`
	Y           []cartesianSeries `json:"y"`
}

// Register adds the bar/line/area chart tool, linking the shared chart widget.
func Register(s *mcp.Server, deps shared.Deps) {
	tool := &mcp.Tool{
		Name:        "show_bar_line_area_chart",
		Description: "Use this for bar charts, line charts, area charts or any combination of them. Shows a chart visualization widget for a result_key received from query_database. All chart types support stacking. Needs one X axis column and one or more Y series. Pass the result_key from query_database plus the axis/series mapping. The chart widget loads the rows itself using the result_key. Note: users can view original SQL in the rendered chart widget, also the table of data which they can download as JSON or CSV.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}
	if uri := ui.WidgetURI(shared.ChartWidget); uri != "" {
		tool.Meta = mcp.Meta{"ui": map[string]any{"resourceUri": uri}}
	}
	tool.InputSchema = shared.InputSchema[input](schemaJSON)
	mcp.AddTool(s, tool, handle(deps))
}

// handle renders the stored query as a bar/line/area chart. Like the donut demo
// the structured result is only a discriminator: the linked widget reads this
// tool's arguments and loads its own data via the app-only fetch_chart_data.
func handle(deps shared.Deps) mcp.ToolHandlerFor[input, shared.ChartRenderOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in input) (*mcp.CallToolResult, shared.ChartRenderOutput, error) {
		if len(in.Y) == 0 {
			log.Warn().Str("query_key", in.QueryKey).Msg("chart validation failed: no y series")
			return nil, shared.ChartRenderOutput{}, fmt.Errorf("at least one y series is required")
		}
		if err := deps.RequireQuery(ctx, in.QueryKey); err != nil {
			log.Warn().Str("query_key", in.QueryKey).Err(err).Msg("chart query_key not found")
			return nil, shared.ChartRenderOutput{}, err
		}
		return shared.ChartRenderResult("show_bar_line_area_chart", "bar/line/area chart")
	}
}

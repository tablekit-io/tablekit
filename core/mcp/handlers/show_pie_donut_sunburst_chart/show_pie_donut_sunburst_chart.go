// Package showpiedonutsunburstchart implements the show_pie_donut_sunburst_chart MCP tool.
package showpiedonutsunburstchart

import (
	"context"
	_ "embed"
	"fmt"

	"core/helpers"
	"core/mcp/handlers/shared"
	"core/mcp/ui"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed schema.json
var schemaJSON []byte

// proportionalLayer is one grouping ring: each distinct value of the column
// becomes a slice, with values summed per group. Layers go innermost first.
type proportionalLayer struct {
	DiscriminatorProp string `json:"discriminator_prop"`
}

// input is the show_pie_donut_sunburst_chart tool's argument schema. Descriptions
// and the display enum live in schema.json; the struct only decodes.
type input struct {
	QueryKey    string              `json:"query_key"`
	Display     string              `json:"display,omitempty"`
	ValueProp   string              `json:"value_prop"`
	ValuePrefix string              `json:"value_prefix,omitempty"`
	ValueSuffix string              `json:"value_suffix,omitempty"`
	Layers      []proportionalLayer `json:"layers"`
}

// Register adds the proportional chart tool, linking the shared chart widget.
func Register(s *mcp.Server, deps shared.Deps) {
	tool := &mcp.Tool{
		Name:        "show_pie_donut_sunburst_chart",
		Description: "Use this for donut or pie charts. Shows a proportional chart visualization widget for a result_key received from query_database. Both chart types support stacking, stacking will result in a sunburst chart. Needs the value column and one or more grouping layers (inner-most ring first). Pass the result_key from query_database along with the columns & grouping. The chart widget loads the rows itself using the result_key. Note: users can view original SQL in the rendered chart widget, also the table of data which they can download as JSON or CSV.",
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

// handle renders the stored query as a proportional (pie/donut/sunburst) chart.
// The structured result is only a discriminator; the linked widget reads this
// tool's arguments and loads rows via the app-only fetch_chart_data.
func handle(deps shared.Deps) mcp.ToolHandlerFor[input, shared.ChartRenderOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in input) (*mcp.CallToolResult, shared.ChartRenderOutput, error) {
		if in.ValueProp == "" {
			return nil, shared.ChartRenderOutput{}, fmt.Errorf("value_prop is required")
		}
		if len(in.Layers) == 0 {
			return nil, shared.ChartRenderOutput{}, fmt.Errorf("at least one layer is required")
		}
		if err := deps.RequireQuery(ctx, in.QueryKey); err != nil {
			return nil, shared.ChartRenderOutput{}, err
		}
		return shared.ChartRenderResult("show_pie_donut_sunburst_chart", "pie/donut/sunburst chart")
	}
}

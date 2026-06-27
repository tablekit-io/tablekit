package handlers

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// dataInput is the hello_world_interactive_data loader's argument schema.
type dataInput struct {
	Slices int `json:"slices,omitempty" jsonschema:"number of donut slices (1-8); defaults to 5"`
}

// dataSlice is one donut slice. Lowercase json tags match what the widget reads
// off structuredContent.
type dataSlice struct {
	Label string  `json:"label" jsonschema:"the slice label"`
	Value float64 `json:"value" jsonschema:"the slice value"`
}

// dataOutput is the loader's structured result: the random slices to plot.
type dataOutput struct {
	Data []dataSlice `json:"data" jsonschema:"the donut slices"`
}

// helloInteractiveData is the example data loader: it returns a random dataset
// for the donut. App-only — hidden from the model, called only by the widget
// over the MCP Apps bridge.
func (h *Handlers) helloInteractiveData(_ context.Context, _ *mcp.CallToolRequest, in dataInput) (*mcp.CallToolResult, dataOutput, error) {
	n := in.Slices
	if n < 1 {
		n = 5
	}
	if n > 8 {
		n = 8
	}
	slices := make([]dataSlice, n)
	for i := 0; i < n; i++ {
		slices[i] = dataSlice{
			Label: fmt.Sprintf("Category %c", 'A'+i),
			Value: float64(rand.Intn(91) + 10), // 10..100
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Generated %d random donut slices.", n),
		}},
	}, dataOutput{Data: slices}, nil
}

// registerHelloWorldInteractiveData adds the example data loader. The
// _meta.ui.visibility=['app'] marks it app-only — the host hides it from the
// model and only honours it when the widget calls it over the bridge.
func (h *Handlers) registerHelloWorldInteractiveData(s *mcp.Server) {
	dataTool := &mcp.Tool{
		Name:        "hello_world_interactive_data",
		Description: "Returns random example data for the hello_world_interactive donut. App-only: called by the widget over the MCP Apps bridge, hidden from the agent.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: pointer(false),
			OpenWorldHint:   pointer(false),
		},
	}
	dataTool.Meta = mcp.Meta{"ui": map[string]any{"visibility": []string{"app"}}}
	mcp.AddTool(s, dataTool, h.helloInteractiveData)
}

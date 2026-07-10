package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"core/engine"
	"core/mcp/ui"
	"core/services"
	"core/services/config"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectInMemory wires a client session to newServer() over the SDK's
// in-memory transport — exercises the MCP protocol with no HTTP or auth.
func connectInMemory(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverT, clientT := mcp.NewInMemoryTransports()

	// Registration only stores the service handles; the protocol tests below list
	// tools and call list_available_databases. list_available_databases needs a
	// (possibly empty) engine, so wire one loaded from a nonexistent file (yields an
	// empty Service, no error). Queries/Issuer stay nil — it doesn't touch them.
	emptyEngine, err := engine.Load(filepath.Join(t.TempDir(), "none.yaml"), engine.Limits{})
	require.NoError(t, err)
	server := newServer(&services.Services{Config: &config.Config{}, Engine: emptyEngine})
	ss, err := server.Connect(ctx, serverT, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ss.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	clientSession, err := client.Connect(ctx, clientT, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })
	return clientSession
}

// toolsByName lists the server's tools and indexes them by name.
func toolsByName(t *testing.T, clientSession *mcp.ClientSession) map[string]*mcp.Tool {
	t.Helper()
	result, err := clientSession.ListTools(context.Background(), &mcp.ListToolsParams{})
	require.NoError(t, err)
	byName := make(map[string]*mcp.Tool, len(result.Tools))
	for _, tool := range result.Tools {
		byName[tool.Name] = tool
	}
	return byName
}

func TestListToolsExposesAnnotationsAndSchema(t *testing.T) {
	clientSession := connectInMemory(t)

	tool := toolsByName(t, clientSession)["list_available_databases"]
	require.NotNil(t, tool)
	assert.NotNil(t, tool.OutputSchema, "tool should advertise an output schema")

	require.NotNil(t, tool.Annotations)
	assert.True(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, tool.Annotations.IdempotentHint)
	require.NotNil(t, tool.Annotations.DestructiveHint)
	assert.False(t, *tool.Annotations.DestructiveHint)
	require.NotNil(t, tool.Annotations.OpenWorldHint)
	assert.False(t, *tool.Annotations.OpenWorldHint)
}

// uiMeta extracts the _meta.ui map a tool advertises, or nil if absent.
func uiMeta(tool *mcp.Tool) map[string]any {
	ui, ok := tool.Meta["ui"].(map[string]any)
	if !ok {
		return nil
	}
	return ui
}

func TestStoredQueryToolsRegistered(t *testing.T) {
	clientSession := connectInMemory(t)
	tools := toolsByName(t, clientSession)

	for _, name := range []string{
		"query_database", "read_results", "fetch_chart_data",
		"show_bar_line_area_chart", "show_pie_donut_sunburst_chart", "get_export_url",
	} {
		require.NotNil(t, tools[name], "tool %q should be registered", name)
	}

	// fetch_chart_data and get_export_url are app-only: they carry
	// _meta.ui.visibility=['app'] so the host hides them from the model and only
	// honours them when the chart widget calls them over the bridge.
	for _, name := range []string{"fetch_chart_data", "get_export_url"} {
		appUI := uiMeta(tools[name])
		require.NotNil(t, appUI, "%s should carry _meta.ui", name)
		assert.Equal(t, []any{"app"}, appUI["visibility"], "%s should be app-only", name)
	}

	// Build-dependent: the render tools link the shared chart widget via
	// _meta.ui.resourceUri only once @tablekit/widgets has been built into the
	// embed dir.
	if ui.WidgetURI("chart_renderer") != "" {
		for _, name := range []string{"show_bar_line_area_chart", "show_pie_donut_sunburst_chart"} {
			meta := uiMeta(tools[name])
			require.NotNil(t, meta, "built %q should carry _meta.ui", name)
			uri, _ := meta["resourceUri"].(string)
			assert.Contains(t, uri, "ui://tablekit/chart_renderer-")
		}
	}
}

// TestChartToolsAdvertiseEnums confirms the hand-written schema files reach the
// wire: the enum constraints the Go structs can't express (they are plain
// strings) show up in the advertised input schema clients see.
func TestChartToolsAdvertiseEnums(t *testing.T) {
	clientSession := connectInMemory(t)
	tools := toolsByName(t, clientSession)

	// show_bar_line_area_chart: y[].display_as and y[].shape.
	barSchema := marshalToMap(t, tools["show_bar_line_area_chart"].InputSchema)
	series := barSchema["properties"].(map[string]any)["y"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
	assert.ElementsMatch(t, []any{"line", "area", "bar"}, series["display_as"].(map[string]any)["enum"])
	assert.ElementsMatch(t, []any{"line", "discrete", "curve"}, series["shape"].(map[string]any)["enum"])
	assert.NotContains(t, series, "color_hue")

	// show_pie_donut_sunburst_chart: display.
	pieSchema := marshalToMap(t, tools["show_pie_donut_sunburst_chart"].InputSchema)
	display := pieSchema["properties"].(map[string]any)["display"].(map[string]any)
	assert.ElementsMatch(t, []any{"pie", "donut"}, display["enum"])
	assert.Equal(t, "donut", display["default"])
}

// TestChartToolsAdvertiseValuePrefixSuffix asserts the value_prefix/value_suffix
// formatting fields reach the wire: the donut tool's top-level pair, and the
// cartesian tool's chart-root fallback plus its per-series overrides.
func TestChartToolsAdvertiseValuePrefixSuffix(t *testing.T) {
	clientSession := connectInMemory(t)
	tools := toolsByName(t, clientSession)

	// show_pie_donut_sunburst_chart: top-level value_prefix/value_suffix.
	pieProps := marshalToMap(t, tools["show_pie_donut_sunburst_chart"].InputSchema)["properties"].(map[string]any)
	assert.Equal(t, "string", pieProps["value_prefix"].(map[string]any)["type"])
	assert.Equal(t, "string", pieProps["value_suffix"].(map[string]any)["type"])

	// show_bar_line_area_chart: chart-root fallback and per-series overrides.
	barProps := marshalToMap(t, tools["show_bar_line_area_chart"].InputSchema)["properties"].(map[string]any)
	assert.Equal(t, "string", barProps["value_prefix"].(map[string]any)["type"])
	assert.Equal(t, "string", barProps["value_suffix"].(map[string]any)["type"])

	series := barProps["y"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
	assert.Equal(t, "string", series["value_prefix"].(map[string]any)["type"])
	assert.Equal(t, "string", series["value_suffix"].(map[string]any)["type"])
}

// marshalToMap round-trips a tool's InputSchema (whatever concrete type the SDK
// used) through JSON into a generic map for structural assertions.
func marshalToMap(t *testing.T, schema any) map[string]any {
	t.Helper()
	require.NotNil(t, schema)
	raw, err := json.Marshal(schema)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestWidgetResourceIsServed(t *testing.T) {
	// Only meaningful once the widget is built into the embed dir; a fresh
	// checkout ships the placeholder manifest and serves no UI resources.
	resources := ui.Resources()
	if len(resources) == 0 {
		t.Skip("no built widgets in embed dir (run `bun run build` in widgets/)")
	}

	clientSession := connectInMemory(t)
	list, err := clientSession.ListResources(context.Background(), &mcp.ListResourcesParams{})
	require.NoError(t, err)

	var uri string
	for _, r := range list.Resources {
		if r.Name == "chart_renderer" {
			uri = r.URI
			assert.Equal(t, "text/html;profile=mcp-app", r.MIMEType)
		}
	}
	require.NotEmpty(t, uri, "widget resource should be listed")

	read, err := clientSession.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uri})
	require.NoError(t, err)
	require.Len(t, read.Contents, 1)
	assert.Contains(t, read.Contents[0].Text, "<html", "resource should serve the widget HTML")
}

func TestCallToolReturnsStructuredContent(t *testing.T) {
	clientSession := connectInMemory(t)

	// list_available_databases runs against the empty engine wired in
	// connectInMemory, so it returns zero databases — enough to exercise the text +
	// structured result.
	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "list_available_databases",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)

	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "0 database(s) configured.", text.Text)

	structured, ok := result.StructuredContent.(map[string]any)
	require.True(t, ok)
	assert.Empty(t, structured["databases"])
}

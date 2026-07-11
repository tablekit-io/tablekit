// Package listavailabledatabases implements the list_available_databases MCP tool.
package listavailabledatabases

import (
	"context"
	_ "embed"

	"core/engine"
	"core/engine/config"
	"core/helpers"
	"core/mcp/handlers/shared"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed schema.json
var schemaJSON []byte

//go:embed output.tmpl
var outputTmpl []byte

// textTemplate renders the canonical structured output into the text content
// block. Compiled once at init; a malformed template panics at startup.
var textTemplate = shared.MustTemplate(outputTmpl)

// input is empty: the tool takes no arguments.
type input struct{}

// output is the set of configured databases, name and type only.
type output struct {
	Databases        []engine.DatabaseInfo `json:"databases" jsonschema:"lists the databases available to you, includes name and database type (like postgres, mysql, mariadb, bigquery, etc)"`
	HintsForAIAgents []string              `json:"hints_for_ai_agents,omitempty" jsonschema:"guidance for the calling AI agent on how best to use this result"`
}

// bigQueryCostHint is added to the result when any listed database is BigQuery,
// which bills by bytes scanned. It nudges the assistant to explore cheaply before
// committing to a wide query.
const bigQueryCostHint = "One or more databases are BigQuery, which bills by bytes scanned. Before the final query, explore cheaply: inspect INFORMATION_SCHEMA, sample with a tight LIMIT, and test your assumptions with small queries. Then select only the columns and partitions you need — avoid SELECT * on large tables."

// Register adds the list_available_databases tool to s.
func Register(s *mcp.Server, deps shared.Deps) {
	tool := &mcp.Tool{
		Name:        "list_available_databases",
		Description: "Lists the databases that query_database can run against, each with its name and engine type (postgres/mysql/mariadb/bigquery). Does not return credentials nor connection details.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}
	tool.InputSchema = shared.InputSchema[input](schemaJSON)
	mcp.AddTool(s, tool, handle(deps))
}

// handle returns the databases query_database can run against. It never exposes
// hosts, credentials or any connection detail.
func handle(deps shared.Deps) mcp.ToolHandlerFor[input, output] {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ input) (*mcp.CallToolResult, output, error) {
		out := output{Databases: deps.Engine.List()}
		for _, database := range out.Databases {
			if database.Type == string(config.DatabaseTypeBigQuery) {
				out.HintsForAIAgents = append(out.HintsForAIAgents, bigQueryCostHint)
				break
			}
		}
		text, err := shared.RenderText(textTemplate, out)
		if err != nil {
			return nil, out, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, out, nil
	}
}

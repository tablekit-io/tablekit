// Package listavailabledatabases implements the list_available_databases MCP tool.
package listavailabledatabases

import (
	"context"
	_ "embed"
	"fmt"

	"core/engine"
	"core/helpers"
	"core/mcp/handlers/shared"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed schema.json
var schemaJSON []byte

// input is empty: the tool takes no arguments.
type input struct{}

// output is the set of configured databases, name and type only.
type output struct {
	Databases        []engine.DatabaseInfo `json:"databases" jsonschema:"lists the databases available to you, includes name and database type (like postgres, mysql, mariadb, etc)"`
	HintsForAIAgents []string              `json:"hints_for_ai_agents,omitempty" jsonschema:"guidance for the calling AI agent on how best to use this result"`
}

// Register adds the list_available_databases tool to s.
func Register(s *mcp.Server, engineService *engine.Service) {
	tool := &mcp.Tool{
		Name:        "list_available_databases",
		Description: "Lists the databases that query_database can run against, each with its name and engine type (postgres/mysql/mariadb). Does not return credentials nor connection details.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}
	tool.InputSchema = shared.InputSchema[input](schemaJSON)
	mcp.AddTool(s, tool, handle(engineService))
}

// handle returns the databases query_database can run against. It never exposes
// hosts, credentials or any connection detail.
func handle(engineService *engine.Service) mcp.ToolHandlerFor[input, output] {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ input) (*mcp.CallToolResult, output, error) {
		databases := engineService.List()
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("%d database(s) configured.", len(databases)),
			}},
		}, output{Databases: databases}, nil
	}
}

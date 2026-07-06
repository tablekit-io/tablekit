package handlers

import (
	"context"
	"fmt"

	"core/engine"
	"core/helpers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// listAvailableDatabasesInput is empty: the tool takes no arguments.
type listAvailableDatabasesInput struct{}

// listAvailableDatabasesOutput is the set of configured databases, name and type only.
type listAvailableDatabasesOutput struct {
	Databases        []engine.DatabaseInfo `json:"databases" jsonschema:"lists the databases available to you, includes name and database type (like postgres, mysql, mariadb, etc)"`
	HintsForAIAgents []string              `json:"hints_for_ai_agents,omitempty" jsonschema:"guidance for the calling AI agent on how best to use this result"`
}

// listAvailableDatabases returns the databases query_database can run against. It
// never exposes hosts, credentials or any connection detail.
func (h *Handlers) listAvailableDatabases(_ context.Context, _ *mcp.CallToolRequest, _ listAvailableDatabasesInput) (*mcp.CallToolResult, listAvailableDatabasesOutput, error) {
	databases := h.Engine.List()
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("%d database(s) configured.", len(databases)),
		}},
	}, listAvailableDatabasesOutput{Databases: databases}, nil
}

// registerListAvailableDatabases adds the list_available_databases tool.
func (h *Handlers) registerListAvailableDatabases(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_available_databases",
		Description: "Lists the databases that query_database can run against, each with its name and engine type (postgres/mysql/mariadb). Does not return credentials nor connection details.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}, h.listAvailableDatabases)
}

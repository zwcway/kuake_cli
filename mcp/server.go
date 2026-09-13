package main

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/zwcway/kuake_cli/internal/guard"
	"github.com/zwcway/kuake_cli/mcp/tools"
	"github.com/zwcway/kuake_cli/sdk"
)

func newMCPServer(client *sdk.QuarkClient, g *guard.Guard) *server.MCPServer {
	s := server.NewMCPServer("kuake-mcp", "1.0.0")
	for _, entry := range tools.AllTools(client, g) {
		s.AddTool(entry.Tool, entry.Handler)
	}
	return s
}

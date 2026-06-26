package mcp

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"memory-brain/internal/database"
	"memory-brain/internal/service"
)

type Server struct {
	mcpServer *mcp.Server
	memorySvc *service.MemoryService
	db        *database.Database
}

// NewServer creates a new instance of the Memory Brain MCP Server.
func NewServer(memorySvc *service.MemoryService, db *database.Database) *Server {
	s := &Server{
		memorySvc: memorySvc,
		db:        db,
	}

	// Initialize the MCP server
	s.mcpServer = mcp.NewServer(
		&mcp.Implementation{
			Name:    "memory-brain",
			Version: "1.0.0",
		},
		nil,
	)

	// Register tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.search",
		Description: "Search memories semantically by query text. Reuses the exact core matching logic.",
	}, s.handleSearch)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.remember",
		Description: "Store a new memory (context, fact, decision, troubleshooting case, etc.) semantically.",
	}, s.handleRemember)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.health",
		Description: "Check the status and health of the Memory Brain backend and its database connection.",
	}, s.handleHealth)

	return s
}

// Run executes the MCP server using Stdio transport.
func (s *Server) Run(ctx context.Context) error {
	slog.Info("Running MCP server over Stdio transport...")
	// Run the server using Stdio transport
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

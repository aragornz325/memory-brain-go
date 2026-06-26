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

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.workspace_context",
		Description: "Retrieve operational context for a workspace (stack, conventions, architecture decisions).",
	}, s.handleWorkspaceContext)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.project_snapshot",
		Description: "Retrieve project snapshot (technologies, active features, open decisions, risks, onboarding info).",
	}, s.handleProjectSnapshot)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.related",
		Description: "Fetch memories related to the specified memory ID.",
	}, s.handleRelated)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.decide_save",
		Description: "Automatically analyze if a memory summary should be saved to the database, suggesting workspace, tags, and memory type.",
	}, s.handleDecideSave)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.stats",
		Description: "Retrieve statistics and metrics on database memories and performance.",
	}, s.handleStats)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.link",
		Description: "Link two memories together with a relationship type (e.g., related_to, supersedes).",
	}, s.handleLink)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.bootstrap",
		Description: "Perform session bootstrapping: compiles workspace context, project snapshots, recent decisions, known risks, task-relevant memories, recommended tags, and performance stats.",
	}, s.handleBootstrap)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.detect_workspace",
		Description: "Determine the current active workspace by evaluating matching scores of directory path components against workspace slugs and aliases.",
	}, s.handleDetectWorkspace)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.smart_bootstrap",
		Description: "Automatically detect the current active workspace from the path and execute session bootstrapping for the given task.",
	}, s.handleSmartBootstrap)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "memory.set_workspace_aliases",
		Description: "Set or update workspace aliases (directory naming patterns) for auto workspace detection.",
	}, s.handleSetWorkspaceAliases)

	// Register Resources and Prompts
	s.registerResourcesAndPrompts()

	return s
}

// Run executes the MCP server using Stdio transport.
func (s *Server) Run(ctx context.Context) error {
	slog.Info("Running MCP server over Stdio transport...")
	// Run the server using Stdio transport
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

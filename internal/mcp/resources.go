package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerResourcesAndPrompts() {
	// 1. Resources
	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://rules",
		Name:        "rules",
		Title:       "Rules and Conventions for Agents",
		Description: "General rules, coding guidelines, and practices that all agents working on this workspace should follow.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://save-policy",
		Name:        "save-policy",
		Title:       "Memory Saving Policy",
		Description: "Rules on when to save a memory, what types to classify under, and confirmation requirements.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://workspace-policy",
		Name:        "workspace-policy",
		Title:       "Workspace Management Policy",
		Description: "Policy detailing how workspaces are named, created, and scoped to ensure context separation.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://tagging-conventions",
		Name:        "tagging-conventions",
		Title:       "Tagging Conventions",
		Description: "Standardized conventions for tags (lower-case, kebab-case) to organize semantic memories.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://workspaces",
		Name:        "workspaces",
		Title:       "Available Workspaces",
		Description: "Lists all currently active workspaces and their metadata.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://active-workspaces",
		Name:        "active-workspaces",
		Title:       "Active Workspaces",
		Description: "Lists all currently active workspaces and their metadata.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://workspace-aliases",
		Name:        "workspace-aliases",
		Title:       "Workspace Aliases",
		Description: "Lists all workspaces and their configured directory path aliases.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://bootstrap-guide",
		Name:        "bootstrap-guide",
		Title:       "Session Bootstrap Guide",
		Description: "Onboarding guide detailing how to use the workspace auto-detection and session bootstrap tools.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://server-info",
		Name:        "server-info",
		Title:       "Memory Brain Server Status",
		Description: "Metadata about the Memory Brain MCP server, active LLM model, database statistics, etc.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "memory://deployment-topology",
		Name:        "deployment-topology",
		Title:       "Deployment Topology",
		Description: "Overview of the deployment architecture, host names, LAN connections, and remote setup.",
		MIMEType:    "text/markdown",
	}, s.handleReadResource)

	// 2. Prompts
	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "memory-save-protocol",
		Title:       "Memory Saving Protocol",
		Description: "Guidance on how and when to save context to Memory Brain",
	}, s.handleGetPrompt)

	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "memory-search-protocol",
		Title:       "Memory Search Protocol",
		Description: "Guidance on how to query and locate relevant memories semantically",
	}, s.handleGetPrompt)

	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "workspace-bootstrap",
		Title:       "Workspace Onboarding Bootstrap",
		Description: "Guides an agent in fetching the necessary context and rules for a given workspace",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "workspace",
				Title:       "Workspace Slug",
				Description: "The slug of the workspace to initialize context for (e.g. personal-lab)",
				Required:    true,
			},
		},
	}, s.handleGetPrompt)

	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "architecture-review",
		Title:       "Architecture Review Protocol",
		Description: "Instructions on reviewing code design and architecture aligned with stored decisions",
	}, s.handleGetPrompt)
}

func (s *Server) handleReadResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	var text string

	switch uri {
	case "memory://rules":
		text = `# Memory Brain Rules and Conventions
1. **Always search before starting**: Query Memory Brain semantically before initiating any task.
2. **Document Architecture**: Major architectural changes must be registered under the 'architecture' memory type with an importance of 10.
3. **Link Related Memories**: When a decision updates or refines another, link them using 'supersedes' or 'related_to' to maintain a coherent timeline.
4. **Follow Naming Conventions**: Keep slugs and tags lowercase and kebab-case.
5. **No Duplication**: Check if a similar memory exists before creating a new one.
`
	case "memory://save-policy":
		text = `# Memory Saving Policy
- **Nunca guardar automáticamente**: No persistir datos de forma autónoma.
- **Solicitar confirmación**: Solicitar aprobación explícita del desarrollador antes de guardar.
- **Conocimiento reutilizable**: Guardar únicamente lecciones, decisiones o arquitecturas reutilizables.
- **Evitar datos efímeros**: No guardar logs temporales, comandos de una sola vez, o variables transitorias.
- **Verificar conectividad**: Comprobar el estado mediante 'memory.health' antes de ejecutar 'memory.remember'.
- **Priorizar**: Dar prioridad a decisiones arquitectónicas, lecciones aprendidas y convenciones de equipo.
`
	case "memory://workspace-policy":
		text = `# Workspace Scoping Policy
1. Memories are scoped strictly to Workspaces.
2. Workspaces represent logical business boundaries (e.g. a monorepo, a client project, or a lab environment).
3. The default workspace is 'personal-lab'.
4. Do not pollute workspaces with unrelated domain context.
`
	case "memory://tagging-conventions":
		text = `# Tagging Conventions
- Use only lowercase letters, numbers, and dashes (kebab-case).
- Tag by:
  - Language: 'go', 'typescript', 'python'
  - Framework: 'chi', 'react', 'nestjs'
  - Layer: 'db', 'api', 'domain', 'mcp'
  - Intent: 'bugfix', 'architecture', 'convention'
- Keep tags concise and highly reuse-oriented.
`
	case "memory://workspaces", "memory://active-workspaces":
		workspaces, err := s.memorySvc.ListWorkspaces(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load workspaces for resource: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("# Active Workspaces\n\n")
		for _, ws := range workspaces {
			name := ws.Slug
			if ws.Name != nil {
				name = *ws.Name
			}
			sb.WriteString(fmt.Sprintf("- **%s** (slug: `%s`, ID: `%s`)\n", name, ws.Slug, ws.ID))
		}
		text = sb.String()

	case "memory://workspace-aliases":
		workspaces, err := s.memorySvc.ListWorkspaces(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load workspace aliases for resource: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("# Workspace Aliases\n\n")
		for _, ws := range workspaces {
			wsName := ws.Slug
			if ws.Name != nil && *ws.Name != "" {
				wsName = *ws.Name
			}
			aliasesStr := "[]"
			if len(ws.Aliases) > 0 {
				var formatted []string
				for _, a := range ws.Aliases {
					formatted = append(formatted, fmt.Sprintf("`%s`", a))
				}
				aliasesStr = "[" + strings.Join(formatted, ", ") + "]"
			}
			sb.WriteString(fmt.Sprintf("- **%s** (slug: `%s`): %s\n", wsName, ws.Slug, aliasesStr))
		}
		text = sb.String()

	case "memory://bootstrap-guide":
		text = `# Session Bootstrap Guide
This guide outlines how to onboard yourself and synchronize with the workspace context.

## Auto-detection
If you only have a local directory path, call ` + "`" + `memory.detect_workspace` + "`" + ` to identify the correct workspace slug.

## Bootstrap Execution
Once you have the workspace slug (e.g. ` + "`" + `personal-lab` + "`" + `), run ` + "`" + `memory.bootstrap` + "`" + ` providing the description of the task you are about to do. This will fetch:
- Context rules and project snapshot
- Stored architectural decisions and active risks
- Relevant past memories associated with the task
- LLM-recommended tags for the task

## Smart Bootstrap
Alternatively, you can just call ` + "`" + `memory.smart_bootstrap` + "`" + ` with the directory path and task description directly, and it will do both auto-detection and bootstrapping in a single call.
`

	case "memory://server-info":
		stats, err := s.memorySvc.GetStats(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve stats for resource: %w", err)
		}
		var sb strings.Builder
		sb.WriteString("# Memory Brain Server Information\n\n")
		sb.WriteString(fmt.Sprintf("- **Total Memories:** %d\n", stats.TotalMemories))
		sb.WriteString(fmt.Sprintf("- **Active Workspaces:** %d\n", stats.TotalWorkspaces))
		sb.WriteString(fmt.Sprintf("- **Embedding Model:** %s\n", stats.EmbeddingModel))
		sb.WriteString(fmt.Sprintf("- **Vector Index Size:** %d items\n", stats.VectorIndexSize))
		sb.WriteString(fmt.Sprintf("- **Avg Search Time:** %.2f ms\n", stats.AvgSearchTimeMs))
		sb.WriteString(fmt.Sprintf("- **Avg Embedding Time:** %.2f ms\n", stats.AvgEmbeddingTimeMs))
		sb.WriteString(fmt.Sprintf("- **Avg Bootstrap Time:** %.2f ms\n", stats.AvgBootstrapTimeMs))
		sb.WriteString(fmt.Sprintf("- **Cache Hit Ratio:** %.2f%%\n", stats.CacheHitRatio*100))
		sb.WriteString(fmt.Sprintf("- **Detection Accuracy:** %.2f%%\n", stats.DetectionAccuracy*100))
		text = sb.String()

	case "memory://deployment-topology":
		text = `# Deployment Topology
- **Developer Workstation**: Hosts Cursor and Antigravity, calling the remote MCP Server.
- **Memory Brain MCP Server (Titan - 192.168.1.120)**: Remotely accessible service within LAN.
- **Memory Brain API**: Backend router connecting memory services.
- **Database (192.168.1.120:5500)**: PostgreSQL container utilizing the pgvector extension.
`
	default:
		return nil, mcp.ResourceNotFoundError(uri)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      uri,
				MIMEType: "text/markdown",
				Text:     text,
			},
		},
	}, nil
}

func (s *Server) handleGetPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	name := req.Params.Name
	args := req.Params.Arguments

	switch name {
	case "memory-save-protocol":
		return &mcp.GetPromptResult{
			Description: "Guidelines on when and how to store context in Memory Brain",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: `You are an AI assistant. Follow the memory-saving protocol when proposing to save context to Memory Brain:
1. Propose memory creation ONLY if the knowledge is reusable across multiple sessions (e.g. an architectural decision, an unusual bugfix, a newly established rule/convention).
2. Check if a similar memory exists before saving to avoid duplication.
3. Classify the memory properly using 'memory_type' (architecture, decision, bugfix, workflow, convention, infrastructure, lesson_learned, configuration, deployment) and tag it correctly.
4. Set an appropriate importance score from 1 (transient command) to 10 (architectural core decision).
5. Always prompt the human developer for confirmation before executing remember.`,
					},
				},
			},
		}, nil

	case "memory-search-protocol":
		return &mcp.GetPromptResult{
			Description: "Guidelines on how to query and search for context in Memory Brain",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: `When starting a new session or beginning a new task:
1. Perform a semantic search on Memory Brain using 'memory.search' to check for any existing conventions, rules, or decisions that might affect your work.
2. Use precise query terms, and optionally filter by tags or memory types if you are looking for specific items (like previous bugfixes or architectural design patterns).
3. Synthesize the retrieved memory context into your active planning.`,
					},
				},
			},
		}, nil

	case "workspace-bootstrap":
		ws := args["workspace"]
		if ws == "" {
			ws = "personal-lab"
		}
		return &mcp.GetPromptResult{
			Description: "Bootstrap instructions for workspace: " + ws,
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: fmt.Sprintf(`You are starting to work on the workspace: "%s".
To onboard yourself and set up your context, execute the following steps automatically:
1. Call 'memory.workspace_context' with workspace="%s" to retrieve technology stack, conventions, and architectural decisions.
2. Call 'memory.project_snapshot' with workspace="%s" to get active features, open decisions, and risks.
3. Read the resource "memory://rules" to align with general workspace rules.
Execute these calls before asking the user for tasks, so you can formulate your answers with sufficient background.`, ws, ws, ws),
					},
				},
			},
		}, nil

	case "architecture-review":
		return &mcp.GetPromptResult{
			Description: "Review architecture decisions",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: `To perform an architecture review:
1. Search Memory Brain for memories with types=["architecture", "decision"].
2. Review the design decisions and constraints of the workspace.
3. Evaluate if the proposed design or current codebase aligns with the documented architecture. If conflicts exist, point them out and reference the superseded decisions.`,
					},
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("prompt not found: %s", name)
	}
}

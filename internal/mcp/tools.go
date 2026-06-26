package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"memory-brain/internal/domain"
	"memory-brain/internal/service"
)

// handleSearch handles semantic searching of memories.
func (s *Server) handleSearch(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
	workspaceSlug := args.WorkspaceSlug
	if workspaceSlug == "" {
		workspaceSlug = os.Getenv("MEMORY_BRAIN_WORKSPACE")
	}
	if workspaceSlug == "" {
		workspaceSlug = "personal-lab" // fallback
	}

	projectSlug := args.ProjectSlug
	if projectSlug == "" {
		projectSlug = os.Getenv("MEMORY_BRAIN_PROJECT")
	}
	if projectSlug == "" {
		projectSlug = "memory-brain" // fallback
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 10
	}

	var projectSlugPtr *string
	if projectSlug != "" {
		projectSlugPtr = &projectSlug
	}

	var memoryTypes []domain.MemoryType
	if len(args.Types) > 0 {
		memoryTypes = make([]domain.MemoryType, len(args.Types))
		for i, t := range args.Types {
			memoryTypes[i] = domain.MemoryType(t)
		}
	}

	results, err := s.memorySvc.Search(ctx, &service.SearchMemoryInput{
		WorkspaceSlug: workspaceSlug,
		ProjectSlug:   projectSlugPtr,
		Query:         args.Query,
		Limit:         &limit,
		Types:         memoryTypes,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %w", err)
	}

	type itemResponse struct {
		ID         string   `json:"id"`
		Text       string   `json:"text"`
		Tags       []string `json:"tags"`
		Source     string   `json:"source"`
		CreatedAt  string   `json:"createdAt"`
		Score      float64  `json:"score"`
		MemoryType string   `json:"memoryType,omitempty"`
		Status     string   `json:"status,omitempty"`
		Importance int      `json:"importance"`
	}

	formattedResults := make([]itemResponse, 0, len(results))
	for _, row := range results {
		source := ""
		if row.Source != nil {
			source = *row.Source
		}
		var mt string
		if row.MemoryType != nil {
			mt = *row.MemoryType
		}
		formattedResults = append(formattedResults, itemResponse{
			ID:         row.ID,
			Text:       row.Content,
			Tags:       row.Tags,
			Source:     source,
			CreatedAt:  row.CreatedAt.Format(time.RFC3339),
			Score:      row.Score,
			MemoryType: mt,
			Status:     row.Status,
			Importance: row.Importance,
		})
	}

	jsonBytes, err := json.MarshalIndent(formattedResults, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format search results: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

// handleRemember stores a new memory semantically.
func (s *Server) handleRemember(ctx context.Context, req *mcp.CallToolRequest, args RememberArgs) (*mcp.CallToolResult, any, error) {
	workspaceSlug := args.WorkspaceSlug
	if workspaceSlug == "" {
		workspaceSlug = os.Getenv("MEMORY_BRAIN_WORKSPACE")
	}
	if workspaceSlug == "" {
		workspaceSlug = "personal-lab"
	}

	projectSlug := args.ProjectSlug
	if projectSlug == "" {
		projectSlug = os.Getenv("MEMORY_BRAIN_PROJECT")
	}
	if projectSlug == "" {
		projectSlug = "memory-brain"
	}

	source := args.Source
	if source == "" {
		source = "mcp"
	}

	var sourceRefPtr *string
	if args.SourceRef != "" {
		sourceRefVal := args.SourceRef
		sourceRefPtr = &sourceRefVal
	}

	var projectSlugPtr *string
	if projectSlug != "" {
		projectSlugPtr = &projectSlug
	}

	var memoryTypePtr *string
	if args.MemoryType != "" {
		mtVal := args.MemoryType
		memoryTypePtr = &mtVal
	}

	var statusPtr *string
	if args.Status != "" {
		stVal := args.Status
		statusPtr = &stVal
	}

	var importancePtr *int
	if args.Importance > 0 {
		impVal := args.Importance
		importancePtr = &impVal
	}

	item, err := s.memorySvc.Remember(ctx, &service.RememberMemoryInput{
		WorkspaceSlug: workspaceSlug,
		ProjectSlug:   projectSlugPtr,
		Text:          args.Text,
		Tags:          args.Tags,
		Source:        &source,
		SourceRef:     sourceRefPtr,
		MemoryType:    memoryTypePtr,
		Status:        statusPtr,
		Importance:    importancePtr,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("remember failed: %w", err)
	}

	type createdResponse struct {
		Success bool   `json:"success"`
		ID      string `json:"id"`
		Title   string `json:"title"`
	}

	resp := createdResponse{
		Success: true,
		ID:      item.ID,
		Title:   item.Title,
	}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

// handleHealth checks if Memory Brain database connection is functional.
func (s *Server) handleHealth(ctx context.Context, req *mcp.CallToolRequest, args HealthArgs) (*mcp.CallToolResult, any, error) {
	dbErr := ""
	if s.db == nil || s.db.Pool == nil {
		dbErr = "database connection pool is uninitialized"
	} else if err := s.db.Pool.Ping(ctx); err != nil {
		dbErr = fmt.Sprintf("database ping failed: %v", err)
	}

	status := "OK"
	if dbErr != "" {
		status = "ERROR"
	}

	type healthResponse struct {
		Status   string `json:"status"`
		Database string `json:"database"`
	}

	dbStatus := "connected"
	if dbErr != "" {
		dbStatus = dbErr
	}

	resp := healthResponse{
		Status:   status,
		Database: dbStatus,
	}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format health response: %w", err)
	}

	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}
	if status == "ERROR" {
		result.IsError = true
	}

	return result, nil, nil
}

func (s *Server) handleWorkspaceContext(ctx context.Context, req *mcp.CallToolRequest, args WorkspaceContextArgs) (*mcp.CallToolResult, any, error) {
	wsSlug := args.Workspace
	if wsSlug == "" {
		wsSlug = os.Getenv("MEMORY_BRAIN_WORKSPACE")
	}
	if wsSlug == "" {
		wsSlug = "personal-lab"
	}

	result, err := s.memorySvc.WorkspaceContext(ctx, wsSlug)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: result,
			},
		},
	}, nil, nil
}

func (s *Server) handleProjectSnapshot(ctx context.Context, req *mcp.CallToolRequest, args ProjectSnapshotArgs) (*mcp.CallToolResult, any, error) {
	wsSlug := args.Workspace
	if wsSlug == "" {
		wsSlug = os.Getenv("MEMORY_BRAIN_WORKSPACE")
	}
	if wsSlug == "" {
		wsSlug = "personal-lab"
	}

	result, err := s.memorySvc.ProjectSnapshot(ctx, wsSlug)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: result,
			},
		},
	}, nil, nil
}

func (s *Server) handleRelated(ctx context.Context, req *mcp.CallToolRequest, args RelatedArgs) (*mcp.CallToolResult, any, error) {
	results, err := s.memorySvc.GetRelatedMemories(ctx, args.MemoryID)
	if err != nil {
		return nil, nil, err
	}

	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format related memories: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

func (s *Server) handleDecideSave(ctx context.Context, req *mcp.CallToolRequest, args DecideSaveArgs) (*mcp.CallToolResult, any, error) {
	decision, err := s.memorySvc.DecideSave(ctx, args.Summary)
	if err != nil {
		return nil, nil, err
	}

	jsonBytes, err := json.MarshalIndent(decision, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format decision: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

func (s *Server) handleStats(ctx context.Context, req *mcp.CallToolRequest, args StatsArgs) (*mcp.CallToolResult, any, error) {
	stats, err := s.memorySvc.GetStats(ctx)
	if err != nil {
		return nil, nil, err
	}

	jsonBytes, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format stats: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

func (s *Server) handleLink(ctx context.Context, req *mcp.CallToolRequest, args LinkArgs) (*mcp.CallToolResult, any, error) {
	err := s.memorySvc.LinkMemories(ctx, args.FromMemoryID, args.ToMemoryID, args.RelationType)
	if err != nil {
		return nil, nil, err
	}

	type linkResponse struct {
		Success bool `json:"success"`
	}
	resp := linkResponse{Success: true}
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

func (s *Server) handleBootstrap(ctx context.Context, req *mcp.CallToolRequest, args BootstrapArgs) (*mcp.CallToolResult, any, error) {
	resp, err := s.memorySvc.Bootstrap(ctx, args.Workspace, args.CurrentTask, args.MaxMemories)
	if err != nil {
		return nil, nil, err
	}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format bootstrap response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

func (s *Server) handleDetectWorkspace(ctx context.Context, req *mcp.CallToolRequest, args DetectWorkspaceArgs) (*mcp.CallToolResult, any, error) {
	resp, err := s.memorySvc.DetectWorkspace(ctx, args.Path)
	if err != nil {
		return nil, nil, err
	}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format detect workspace response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

func (s *Server) handleSmartBootstrap(ctx context.Context, req *mcp.CallToolRequest, args SmartBootstrapArgs) (*mcp.CallToolResult, any, error) {
	resp, err := s.memorySvc.SmartBootstrap(ctx, args.Path, args.CurrentTask, args.MaxMemories)
	if err != nil {
		return nil, nil, err
	}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format smart bootstrap response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

func (s *Server) handleSetWorkspaceAliases(ctx context.Context, req *mcp.CallToolRequest, args SetWorkspaceAliasesArgs) (*mcp.CallToolResult, any, error) {
	err := s.memorySvc.SetWorkspaceAliases(ctx, args.Workspace, args.Aliases)
	if err != nil {
		return nil, nil, err
	}

	type aliasResponse struct {
		Success bool `json:"success"`
	}
	resp := aliasResponse{Success: true}
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil, nil
}

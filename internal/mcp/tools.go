package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

	results, err := s.memorySvc.Search(ctx, &service.SearchMemoryInput{
		WorkspaceSlug: workspaceSlug,
		ProjectSlug:   projectSlugPtr,
		Query:         args.Query,
		Limit:         &limit,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %w", err)
	}

	type itemResponse struct {
		ID        string   `json:"id"`
		Text      string   `json:"text"`
		Tags      []string `json:"tags"`
		Source    string   `json:"source"`
		CreatedAt string   `json:"createdAt"`
		Score     float64  `json:"score"`
	}

	formattedResults := make([]itemResponse, 0, len(results))
	for _, row := range results {
		source := ""
		if row.Source != nil {
			source = *row.Source
		}
		formattedResults = append(formattedResults, itemResponse{
			ID:        row.ID,
			Text:      row.Content,
			Tags:      row.Tags,
			Source:    source,
			CreatedAt: row.CreatedAt.Format(time.RFC3339),
			Score:     row.Score,
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

	item, err := s.memorySvc.Remember(ctx, &service.RememberMemoryInput{
		WorkspaceSlug: workspaceSlug,
		ProjectSlug:   projectSlugPtr,
		Text:          args.Text,
		Tags:          args.Tags,
		Source:        &source,
		SourceRef:     sourceRefPtr,
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

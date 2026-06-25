package http

import (
	"time"

	"memory-brain/internal/domain"
)

type MemoryItemResponse struct {
	ID            string   `json:"id"`
	Text          string   `json:"text"`
	Tags          []string `json:"tags"`
	Source        *string  `json:"source,omitempty"`
	SourceRef     *string  `json:"sourceRef,omitempty"`
	WorkspaceSlug *string  `json:"workspaceSlug,omitempty"`
	ProjectSlug   *string  `json:"projectSlug,omitempty"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
	Score         *float64 `json:"score,omitempty"`
}

func mapToResponse(item *domain.MemoryItem, workspaceSlug string, projectSlug *string) *MemoryItemResponse {
	var wsSlugPtr *string
	if workspaceSlug != "" {
		wsSlugPtr = &workspaceSlug
	}

	return &MemoryItemResponse{
		ID:            item.ID,
		Text:          item.Content,
		Tags:          item.Tags,
		Source:        item.Source,
		SourceRef:     item.SourceRef,
		WorkspaceSlug: wsSlugPtr,
		ProjectSlug:   projectSlug,
		CreatedAt:     item.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:     item.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func mapSearchRowToResponse(row *domain.MemorySearchRow, workspaceSlug string, projectSlug *string) *MemoryItemResponse {
	resp := mapToResponse(&row.MemoryItem, workspaceSlug, projectSlug)
	scoreVal := row.Score
	resp.Score = &scoreVal
	return resp
}

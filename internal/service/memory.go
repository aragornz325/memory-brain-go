package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"memory-brain/internal/domain"
)

type MemoryService struct {
	workspaceRepo  WorkspaceRepository
	projectRepo    ProjectRepository
	memoryItemRepo MemoryItemRepository
	embeddingSvc   *EmbeddingService
}

func NewMemoryService(
	workspaceRepo WorkspaceRepository,
	projectRepo ProjectRepository,
	memoryItemRepo MemoryItemRepository,
	embeddingSvc *EmbeddingService,
) *MemoryService {
	return &MemoryService{
		workspaceRepo:  workspaceRepo,
		projectRepo:    projectRepo,
		memoryItemRepo: memoryItemRepo,
		embeddingSvc:   embeddingSvc,
	}
}

type CreateMemoryInput struct {
	WorkspaceSlug string                 `json:"workspaceSlug" validate:"required"`
	ProjectSlug   *string                `json:"projectSlug,omitempty"`
	Type          domain.MemoryType      `json:"type" validate:"required"`
	Title         string                 `json:"title" validate:"required"`
	Content       string                 `json:"content" validate:"required"`
	Summary       *string                `json:"summary,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Source        *string                `json:"source,omitempty"`
	SourceRef     *string                `json:"sourceRef,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Importance    *int                   `json:"importance,omitempty"`
	Confidence    *float64               `json:"confidence,omitempty"`
	IsActive      *bool                  `json:"isActive,omitempty"`
}

type RememberMemoryInput struct {
	WorkspaceSlug string   `json:"workspaceSlug" validate:"required"`
	ProjectSlug   *string  `json:"projectSlug,omitempty"`
	Text          string   `json:"text" validate:"required"`
	Tags          []string `json:"tags,omitempty"`
	Source        *string  `json:"source,omitempty"`
	SourceRef     *string  `json:"sourceRef,omitempty"`
}

type SearchMemoryInput struct {
	WorkspaceSlug string              `json:"workspaceSlug" validate:"required"`
	ProjectSlug   *string             `json:"projectSlug,omitempty"`
	Query         string              `json:"query" validate:"required"`
	Types         []domain.MemoryType `json:"types,omitempty"`
	Tags          []string            `json:"tags,omitempty"`
	OnlyActive    *bool               `json:"onlyActive,omitempty"`
	Limit         *int                `json:"limit,omitempty"`
}

type UpdateMemoryInput struct {
	Type       *domain.MemoryType     `json:"type,omitempty"`
	Title      *string                `json:"title,omitempty"`
	Content    *string                `json:"content,omitempty"`
	Summary    *string                `json:"summary,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Source     *string                `json:"source,omitempty"`
	SourceRef  *string                `json:"sourceRef,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Importance *int                   `json:"importance,omitempty"`
	Confidence *float64               `json:"confidence,omitempty"`
	IsActive   *bool                  `json:"isActive,omitempty"`
}

type ContextResponse struct {
	Context  string                    `json:"context"`
	Memories []*domain.MemorySearchRow `json:"memories"`
}

func (s *MemoryService) Create(ctx context.Context, input *CreateMemoryInput) (*domain.MemoryItem, error) {
	ws, err := s.workspaceRepo.GetBySlug(ctx, input.WorkspaceSlug)
	if err != nil {
		return nil, fmt.Errorf("workspace %s not found: %w", input.WorkspaceSlug, err)
	}

	var projectID *string
	if input.ProjectSlug != nil && *input.ProjectSlug != "" {
		p, err := s.projectRepo.GetBySlug(ctx, ws.ID, *input.ProjectSlug)
		if err != nil {
			if errors.Is(err, domain.ErrProjectNotFound) {
				// Project doesn't exist under this workspace. Create it automatically.
				p = &domain.Project{
					WorkspaceID: ws.ID,
					Slug:        *input.ProjectSlug,
				}
				if err := s.projectRepo.Create(ctx, p); err != nil {
					return nil, fmt.Errorf("failed to auto-create project %s in workspace %s: %w", *input.ProjectSlug, ws.Slug, err)
				}
			} else {
				return nil, fmt.Errorf("failed to query project %s: %w", *input.ProjectSlug, err)
			}
		}
		projectID = &p.ID
	}

	embeddingText := fmt.Sprintf("%s\n%s", input.Title, input.Content)
	embedding, err := s.embeddingSvc.Embed(ctx, embeddingText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	importance := 0
	if input.Importance != nil {
		importance = *input.Importance
	}

	confidence := 1.0
	if input.Confidence != nil {
		confidence = *input.Confidence
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	item := &domain.MemoryItem{
		WorkspaceID: ws.ID,
		ProjectID:   projectID,
		Type:        input.Type,
		Title:       input.Title,
		Content:     input.Content,
		Summary:     input.Summary,
		Tags:        input.Tags,
		Source:      input.Source,
		SourceRef:   input.SourceRef,
		Metadata:    metadata,
		Importance:  importance,
		Confidence:  confidence,
		IsActive:    isActive,
		Embedding:   embedding,
	}

	if item.Tags == nil {
		item.Tags = []string{}
	}

	if err := s.memoryItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *MemoryService) Remember(ctx context.Context, input *RememberMemoryInput) (*domain.MemoryItem, error) {
	trimmedText := strings.TrimSpace(input.Text)
	title := extractRememberTitle(trimmedText)
	summary := truncateText(trimmedText, 500)

	source := "cursor"
	if input.Source != nil && *input.Source != "" {
		source = *input.Source
	}

	importance := 5
	confidence := 1.0
	isActive := true

	createInput := &CreateMemoryInput{
		WorkspaceSlug: input.WorkspaceSlug,
		ProjectSlug:   input.ProjectSlug,
		Type:          domain.TypeConversation,
		Title:         title,
		Content:       trimmedText,
		Summary:       &summary,
		Tags:          input.Tags,
		Source:        &source,
		SourceRef:     input.SourceRef,
		Importance:    &importance,
		Confidence:    &confidence,
		IsActive:      &isActive,
	}

	return s.Create(ctx, createInput)
}

func (s *MemoryService) FindOne(ctx context.Context, id string) (*domain.MemoryItem, error) {
	item, err := s.memoryItemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *MemoryService) Update(ctx context.Context, id string, input *UpdateMemoryInput) (*domain.MemoryItem, error) {
	item, err := s.memoryItemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	contentChanged := false
	if input.Title != nil && *input.Title != item.Title {
		item.Title = *input.Title
		contentChanged = true
	}
	if input.Content != nil && *input.Content != item.Content {
		item.Content = *input.Content
		contentChanged = true
	}

	if input.Type != nil {
		item.Type = *input.Type
	}
	if input.Summary != nil {
		item.Summary = input.Summary
	}
	if input.Tags != nil {
		item.Tags = input.Tags
	}
	if input.Source != nil {
		item.Source = input.Source
	}
	if input.SourceRef != nil {
		item.SourceRef = input.SourceRef
	}
	if input.Metadata != nil {
		item.Metadata = input.Metadata
	}
	if input.Importance != nil {
		item.Importance = *input.Importance
	}
	if input.Confidence != nil {
		item.Confidence = *input.Confidence
	}
	if input.IsActive != nil {
		item.IsActive = *input.IsActive
	}

	if contentChanged {
		embeddingText := fmt.Sprintf("%s\n%s", item.Title, item.Content)
		embedding, err := s.embeddingSvc.Embed(ctx, embeddingText)
		if err != nil {
			return nil, fmt.Errorf("failed to generate updated embedding: %w", err)
		}
		item.Embedding = embedding
	}

	if err := s.memoryItemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *MemoryService) Remove(ctx context.Context, id string) error {
	item, err := s.memoryItemRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	item.IsActive = false
	return s.memoryItemRepo.Update(ctx, item)
}

func (s *MemoryService) Search(ctx context.Context, input *SearchMemoryInput) ([]*domain.MemorySearchRow, error) {
	ws, err := s.workspaceRepo.GetBySlug(ctx, input.WorkspaceSlug)
	if err != nil {
		return nil, fmt.Errorf("workspace %s not found: %w", input.WorkspaceSlug, err)
	}

	var projectID *string
	if input.ProjectSlug != nil && *input.ProjectSlug != "" {
		p, err := s.projectRepo.GetBySlug(ctx, ws.ID, *input.ProjectSlug)
		if err != nil {
			return nil, fmt.Errorf("project %s not found in workspace %s: %w", *input.ProjectSlug, ws.Slug, err)
		}
		projectID = &p.ID
	}

	limit := 10
	if input.Limit != nil {
		limit = *input.Limit
	}

	onlyActive := true
	if input.OnlyActive != nil {
		onlyActive = *input.OnlyActive
	}

	queryEmbedding, err := s.embeddingSvc.Embed(ctx, input.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	rows, err := s.memoryItemRepo.Search(ctx, ws.ID, projectID, queryEmbedding, input.Types, input.Tags, onlyActive, limit)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *MemoryService) GetContext(ctx context.Context, input *SearchMemoryInput) (*ContextResponse, error) {
	memories, err := s.Search(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(memories) == 0 {
		return &ContextResponse{Context: "", Memories: memories}, nil
	}

	var sb strings.Builder
	for i, m := range memories {
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		sb.WriteString(fmt.Sprintf("## %s", m.Content))
		if len(m.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("\nTags: %s", strings.Join(m.Tags, ", ")))
		}
		sb.WriteString(fmt.Sprintf("\nRelevance: %.4f", m.Score))
	}

	return &ContextResponse{
		Context:  sb.String(),
		Memories: memories,
	}, nil
}

// Helpers
func truncateText(text string, maxLength int) string {
	normalized := strings.TrimSpace(text)
	if len(normalized) <= maxLength {
		return normalized
	}
	if maxLength <= 3 {
		return normalized[:maxLength]
	}
	return normalized[:maxLength-3] + "..."
}

func extractRememberTitle(text string) string {
	lines := strings.Split(text, "\n")
	firstLine := ""
	if len(lines) > 0 {
		firstLine = strings.TrimSpace(lines[0])
	}
	if firstLine == "" {
		return "Untitled memory"
	}
	return truncateText(firstLine, 120)
}

func (s *MemoryService) CreateWorkspace(ctx context.Context, slug, name string) (*domain.Workspace, error) {
	ws := &domain.Workspace{
		Slug: slug,
	}
	if name != "" {
		ws.Name = &name
	}
	if err := s.workspaceRepo.Create(ctx, ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *MemoryService) CreateProject(ctx context.Context, workspaceSlug, slug string) (*domain.Project, error) {
	ws, err := s.workspaceRepo.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return nil, fmt.Errorf("workspace %s not found: %w", workspaceSlug, err)
	}

	p := &domain.Project{
		WorkspaceID: ws.ID,
		Slug:        slug,
	}
	if err := s.projectRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

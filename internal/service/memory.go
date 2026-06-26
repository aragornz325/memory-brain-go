package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"memory-brain/internal/domain"
)

type Metrics struct {
	searchCount       int64
	totalSearchTimeNs int64
	embedCount        int64
	totalEmbedTimeNs  int64

	bootstrapCount       int64
	totalBootstrapTimeNs int64
	cacheHits            int64
	cacheMisses          int64
	detectionSuccess     int64
	totalDetections      int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) RecordSearch(d time.Duration) {
	atomic.AddInt64(&m.searchCount, 1)
	atomic.AddInt64(&m.totalSearchTimeNs, int64(d))
}

func (m *Metrics) RecordEmbed(d time.Duration) {
	atomic.AddInt64(&m.embedCount, 1)
	atomic.AddInt64(&m.totalEmbedTimeNs, int64(d))
}

func (m *Metrics) RecordBootstrap(d time.Duration) {
	atomic.AddInt64(&m.bootstrapCount, 1)
	atomic.AddInt64(&m.totalBootstrapTimeNs, int64(d))
}

func (m *Metrics) RecordCacheHit() {
	atomic.AddInt64(&m.cacheHits, 1)
}

func (m *Metrics) RecordCacheMiss() {
	atomic.AddInt64(&m.cacheMisses, 1)
}

func (m *Metrics) RecordDetection(success bool) {
	atomic.AddInt64(&m.totalDetections, 1)
	if success {
		atomic.AddInt64(&m.detectionSuccess, 1)
	}
}

func (m *Metrics) SearchAverageMs() float64 {
	count := atomic.LoadInt64(&m.searchCount)
	if count == 0 {
		return 0
	}
	total := atomic.LoadInt64(&m.totalSearchTimeNs)
	return float64(total) / float64(count) / float64(time.Millisecond)
}

func (m *Metrics) EmbedAverageMs() float64 {
	count := atomic.LoadInt64(&m.embedCount)
	if count == 0 {
		return 0
	}
	total := atomic.LoadInt64(&m.totalEmbedTimeNs)
	return float64(total) / float64(count) / float64(time.Millisecond)
}

func (m *Metrics) BootstrapAverageMs() float64 {
	count := atomic.LoadInt64(&m.bootstrapCount)
	if count == 0 {
		return 0
	}
	total := atomic.LoadInt64(&m.totalBootstrapTimeNs)
	return float64(total) / float64(count) / float64(time.Millisecond)
}

func (m *Metrics) CacheHitRatio() float64 {
	hits := atomic.LoadInt64(&m.cacheHits)
	misses := atomic.LoadInt64(&m.cacheMisses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total)
}

func (m *Metrics) DetectionAccuracy() float64 {
	total := atomic.LoadInt64(&m.totalDetections)
	if total == 0 {
		return 0
	}
	success := atomic.LoadInt64(&m.detectionSuccess)
	return float64(success) / float64(total)
}

type MemoryService struct {
	workspaceRepo  WorkspaceRepository
	projectRepo    ProjectRepository
	memoryItemRepo MemoryItemRepository
	embeddingSvc   *EmbeddingService
	llmSvc         *LLMService
	memoryLinkRepo MemoryLinkRepository
	metrics        *Metrics
	cache          *Cache
}

func NewMemoryService(
	workspaceRepo WorkspaceRepository,
	projectRepo ProjectRepository,
	memoryItemRepo MemoryItemRepository,
	embeddingSvc *EmbeddingService,
	llmSvc *LLMService,
	memoryLinkRepo MemoryLinkRepository,
) *MemoryService {
	metrics := NewMetrics()
	return &MemoryService{
		workspaceRepo:  workspaceRepo,
		projectRepo:    projectRepo,
		memoryItemRepo: memoryItemRepo,
		embeddingSvc:   embeddingSvc,
		llmSvc:         llmSvc,
		memoryLinkRepo: memoryLinkRepo,
		metrics:        metrics,
		cache:          NewCache(metrics),
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
	MemoryType    *string                `json:"memory_type,omitempty"`
	Status        *string                `json:"status,omitempty"`
}

type RememberMemoryInput struct {
	WorkspaceSlug string   `json:"workspaceSlug" validate:"required"`
	ProjectSlug   *string  `json:"projectSlug,omitempty"`
	Text          string   `json:"text" validate:"required"`
	Tags          []string `json:"tags,omitempty"`
	Source        *string  `json:"source,omitempty"`
	SourceRef     *string  `json:"sourceRef,omitempty"`
	MemoryType    *string  `json:"memoryType,omitempty"`
	Status        *string  `json:"status,omitempty"`
	Importance    *int     `json:"importance,omitempty"`
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
	MemoryType *string                `json:"memory_type,omitempty"`
	Status     *string                `json:"status,omitempty"`
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
	embedStart := time.Now()
	embedding, err := s.embeddingSvc.Embed(ctx, embeddingText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}
	s.metrics.RecordEmbed(time.Since(embedStart))

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
		MemoryType:  input.MemoryType,
	}

	if input.Status != nil {
		item.Status = *input.Status
	}

	if item.Tags == nil {
		item.Tags = []string{}
	}

	if err := s.memoryItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	s.cache.Invalidate(ws.Slug + ":")
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

	var memoryType *string
	if input.MemoryType != nil && *input.MemoryType != "" {
		memoryType = input.MemoryType
	}
	importance := 5
	if input.Importance != nil {
		importance = *input.Importance
	}
	confidence := 1.0
	isActive := true

	// Heuristics for tagging-to-type mapping and importance scoring (only if not explicitly provided)
	if memoryType == nil {
		for _, tag := range input.Tags {
			t := strings.ToLower(tag)
			switch t {
			case "architecture":
				mt := domain.MemoryTypeArchitecture
				memoryType = &mt
				if input.Importance == nil {
					importance = 10
				}
			case "decision":
				mt := domain.MemoryTypeDecision
				memoryType = &mt
				if input.Importance == nil {
					importance = 9
				}
			case "bugfix":
				mt := domain.MemoryTypeBugfix
				memoryType = &mt
				if input.Importance == nil {
					importance = 8
				}
			case "workflow":
				mt := domain.MemoryTypeWorkflow
				memoryType = &mt
				if input.Importance == nil {
					importance = 6
				}
			case "convention":
				mt := domain.MemoryTypeConvention
				memoryType = &mt
				if input.Importance == nil {
					importance = 9
				}
			case "infrastructure":
				mt := domain.MemoryTypeInfrastructure
				memoryType = &mt
				if input.Importance == nil {
					importance = 8
				}
			case "lesson_learned":
				mt := domain.MemoryTypeLessonLearned
				memoryType = &mt
				if input.Importance == nil {
					importance = 7
				}
			case "configuration":
				mt := domain.MemoryTypeConfiguration
				memoryType = &mt
				if input.Importance == nil {
					importance = 5
				}
			case "deployment":
				mt := domain.MemoryTypeDeployment
				memoryType = &mt
				if input.Importance == nil {
					importance = 6
				}
			}
		}
	}

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
		MemoryType:    memoryType,
		Status:        input.Status,
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
	if input.MemoryType != nil {
		item.MemoryType = input.MemoryType
	}
	if input.Status != nil {
		item.Status = *input.Status
	}

	if contentChanged {
		embeddingText := fmt.Sprintf("%s\n%s", item.Title, item.Content)
		embedStart := time.Now()
		embedding, err := s.embeddingSvc.Embed(ctx, embeddingText)
		if err != nil {
			return nil, fmt.Errorf("failed to generate updated embedding: %w", err)
		}
		s.metrics.RecordEmbed(time.Since(embedStart))
		item.Embedding = embedding
	}

	if err := s.memoryItemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	s.cache.Invalidate(item.WorkspaceSlug + ":")
	return item, nil
}

func (s *MemoryService) Remove(ctx context.Context, id string) error {
	item, err := s.memoryItemRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	item.IsActive = false
	if err := s.memoryItemRepo.Update(ctx, item); err != nil {
		return err
	}
	s.cache.Invalidate(item.WorkspaceSlug + ":")
	return nil
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

	embedStart := time.Now()
	queryEmbedding, err := s.embeddingSvc.Embed(ctx, input.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	s.metrics.RecordEmbed(time.Since(embedStart))

	repoStart := time.Now()
	rows, err := s.memoryItemRepo.Search(ctx, ws.ID, projectID, queryEmbedding, input.Types, input.Tags, onlyActive, limit)
	if err != nil {
		return nil, err
	}
	s.metrics.RecordSearch(time.Since(repoStart))

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

// Stats & Obsolescence definitions
type Stats struct {
	TotalMemories      int     `json:"total_memories"`
	TotalWorkspaces    int     `json:"total_workspaces"`
	AvgSearchTimeMs    float64 `json:"avg_search_time_ms"`
	AvgEmbeddingTimeMs float64 `json:"avg_embedding_time_ms"`
	EmbeddingModel     string  `json:"embedding_model"`
	VectorIndexSize    int     `json:"vector_index_size"`
	AvgBootstrapTimeMs float64 `json:"avg_bootstrap_time_ms"`
	CacheHitRatio      float64 `json:"cache_hit_ratio"`
	DetectionAccuracy  float64 `json:"detection_accuracy"`
}

func (s *MemoryService) GetStats(ctx context.Context) (*Stats, error) {
	totalMemories, vectorSize, err := s.memoryItemRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	totalWorkspaces, err := s.workspaceRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	return &Stats{
		TotalMemories:      totalMemories,
		TotalWorkspaces:    totalWorkspaces,
		AvgSearchTimeMs:    s.metrics.SearchAverageMs(),
		AvgEmbeddingTimeMs: s.metrics.EmbedAverageMs(),
		EmbeddingModel:     s.embeddingSvc.model,
		VectorIndexSize:    vectorSize,
		AvgBootstrapTimeMs: s.metrics.BootstrapAverageMs(),
		CacheHitRatio:      s.metrics.CacheHitRatio(),
		DetectionAccuracy:  s.metrics.DetectionAccuracy(),
	}, nil
}

func (s *MemoryService) ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error) {
	return s.workspaceRepo.List(ctx)
}

type DecideSaveResponse struct {
	ShouldSave          bool     `json:"should_save"`
	Confidence          float64  `json:"confidence"`
	SuggestedTags       []string `json:"suggested_tags"`
	SuggestedWorkspace  string   `json:"suggested_workspace"`
	SuggestedMemoryType string   `json:"suggested_memory_type"`
	Reason              string   `json:"reason"`
}

func (s *MemoryService) DecideSave(ctx context.Context, summary string) (*DecideSaveResponse, error) {
	prompt := fmt.Sprintf(`Analyze the following text that represents a potential memory to be saved. Determine:
1. If it should be saved as persistent knowledge (e.g. architecture decisions, bugfixes, workflows, conventions, lessons learned). Avoid ephemeral/temporary info.
2. The confidence level (0.0 to 1.0) of this recommendation.
3. A list of suggested tags (lowercase, kebab-case).
4. Suggested workspace slug (e.g. 'personal-lab').
5. Suggested memory_type (must be one of: architecture, decision, bugfix, workflow, convention, infrastructure, lesson_learned, configuration, deployment).
6. Reason for the decision.

Respond ONLY in JSON format matching this schema:
{
  "should_save": boolean,
  "confidence": number,
  "suggested_tags": [string],
  "suggested_workspace": string,
  "suggested_memory_type": string,
  "reason": string
}

Text to analyze:
%s`, summary)

	respText, err := s.llmSvc.Generate(ctx, prompt, true)
	if err != nil {
		return nil, err
	}

	var decision DecideSaveResponse
	if err := json.Unmarshal([]byte(respText), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse decision JSON: %w. Raw response: %s", err, respText)
	}

	return &decision, nil
}

func (s *MemoryService) LinkMemories(ctx context.Context, fromID, toID, relationType string) error {
	link := &domain.MemoryLink{
		FromMemoryID: fromID,
		ToMemoryID:   toID,
		RelationType: relationType,
	}
	err := s.memoryLinkRepo.Create(ctx, link)
	if err != nil {
		return err
	}

	// Auto obsolescence check
	if relationType == "supersedes" {
		oldMem, err := s.memoryItemRepo.FindByID(ctx, toID)
		if err == nil {
			oldMem.Status = domain.StatusObsolete
			oldMem.IsActive = false
			_ = s.memoryItemRepo.Update(ctx, oldMem)
		}
	} else if relationType == "superseded_by" {
		oldMem, err := s.memoryItemRepo.FindByID(ctx, fromID)
		if err == nil {
			oldMem.Status = domain.StatusObsolete
			oldMem.IsActive = false
			_ = s.memoryItemRepo.Update(ctx, oldMem)
		}
	}

	if fromMem, err := s.memoryItemRepo.FindByID(ctx, fromID); err == nil {
		s.cache.Invalidate(fromMem.WorkspaceSlug + ":")
	}
	if toMem, err := s.memoryItemRepo.FindByID(ctx, toID); err == nil {
		s.cache.Invalidate(toMem.WorkspaceSlug + ":")
	}

	return nil
}

type RelatedMemoryInfo struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	RelationType string `json:"relation_type"`
}

func (s *MemoryService) GetRelatedMemories(ctx context.Context, memoryID string) ([]RelatedMemoryInfo, error) {
	links, err := s.memoryLinkRepo.GetLinksForMemory(ctx, memoryID)
	if err != nil {
		return nil, err
	}

	var results []RelatedMemoryInfo
	for _, link := range links {
		relatedID := link.ToMemoryID
		if relatedID == memoryID {
			relatedID = link.FromMemoryID
		}

		item, err := s.memoryItemRepo.FindByID(ctx, relatedID)
		if err == nil {
			summary := ""
			if item.Summary != nil {
				summary = *item.Summary
			}
			results = append(results, RelatedMemoryInfo{
				ID:           item.ID,
				Title:        item.Title,
				Summary:      summary,
				RelationType: link.RelationType,
			})
		}
	}

	return results, nil
}

func (s *MemoryService) WorkspaceContext(ctx context.Context, workspaceSlug string) (string, error) {
	cacheKey := workspaceSlug + ":context"
	if val, ok := s.cache.Get(cacheKey); ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}

	ws, err := s.workspaceRepo.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return "", fmt.Errorf("workspace %s not found: %w", workspaceSlug, err)
	}

	types := []domain.MemoryType{domain.TypeDecision, "convention", "architecture"}
	items, err := s.memoryItemRepo.ListByWorkspace(ctx, ws.ID, types, 15)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Contexto Operativo del Workspace: %s\n\n", workspaceSlug))

	if len(items) == 0 {
		sb.WriteString("No se encontraron decisiones de arquitectura o convenciones registradas para este workspace.\n")
		sb.WriteString("Puedes registrar nuevas usando `memory.remember` con tags como 'architecture' o 'convention'.\n")
		res := sb.String()
		s.cache.Set(cacheKey, res, 10*time.Minute)
		return res, nil
	}

	// Group by memory_type or type
	groups := make(map[string][]string)
	for _, item := range items {
		mType := "general"
		if item.MemoryType != nil && *item.MemoryType != "" {
			mType = *item.MemoryType
		} else if item.Type != "" {
			mType = strings.ToLower(string(item.Type))
		}

		groupText := fmt.Sprintf("### %s (Importancia: %d)\n%s\n", item.Title, item.Importance, item.Content)
		groups[mType] = append(groups[mType], groupText)
	}

	for gName, gItems := range groups {
		sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(gName)))
		for _, gItem := range gItems {
			sb.WriteString(gItem)
			sb.WriteString("\n")
		}
	}

	res := sb.String()
	s.cache.Set(cacheKey, res, 10*time.Minute)
	return res, nil
}

func (s *MemoryService) ProjectSnapshot(ctx context.Context, workspaceSlug string) (string, error) {
	cacheKey := workspaceSlug + ":snapshot"
	if val, ok := s.cache.Get(cacheKey); ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}

	ws, err := s.workspaceRepo.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return "", fmt.Errorf("workspace %s not found: %w", workspaceSlug, err)
	}

	projects, _ := s.projectRepo.ListByWorkspace(ctx, ws.ID)

	items, err := s.memoryItemRepo.ListByWorkspace(ctx, ws.ID, nil, 10)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Snapshot de Proyectos del Workspace: %s\n\n", workspaceSlug))

	if len(projects) > 0 {
		sb.WriteString("## Proyectos Activos\n")
		for _, p := range projects {
			sb.WriteString(fmt.Sprintf("- `%s`\n", p.Slug))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Cambios Recientes & Lecciones Aprendidas\n\n")
	if len(items) == 0 {
		sb.WriteString("No hay memorias recientes en este workspace.\n")
	} else {
		for _, item := range items {
			mType := "general"
			if item.MemoryType != nil && *item.MemoryType != "" {
				mType = *item.MemoryType
			}
			sb.WriteString(fmt.Sprintf("### [%s] %s (Importancia: %d)\n", strings.ToUpper(mType), item.Title, item.Importance))
			if item.Summary != nil && *item.Summary != "" {
				sb.WriteString(fmt.Sprintf("*Resumen:* %s\n\n", *item.Summary))
			}
			sb.WriteString(fmt.Sprintf("%s\n\n", item.Content))
		}
	}

	res := sb.String()
	s.cache.Set(cacheKey, res, 10*time.Minute)
	return res, nil
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

type WorkspaceDetectionResult struct {
	WorkspaceSlug string  `json:"workspace_slug"`
	Score         float64 `json:"score"`
	Confidence    string  `json:"confidence"`
}

type BootstrapStats struct {
	TotalMemories int  `json:"total_memories"`
	TotalProjects int  `json:"total_projects"`
	CacheHit      bool `json:"cache_hit"`
}

type BootstrapResponse struct {
	WorkspaceSlug    string                    `json:"workspace_slug"`
	WorkspaceName    string                    `json:"workspace_name"`
	WorkspaceContext string                    `json:"workspace_context"`
	ProjectSnapshot  string                    `json:"project_snapshot"`
	RecentDecisions  []*domain.MemoryItem      `json:"recent_decisions"`
	KnownRisks       []*domain.MemoryItem      `json:"known_risks"`
	RelevantMemories []*domain.MemorySearchRow `json:"relevant_memories"`
	RecommendedTags  []string                  `json:"recommended_tags"`
	Stats            BootstrapStats            `json:"stats"`
}

func (s *MemoryService) SetWorkspaceAliases(ctx context.Context, workspaceSlug string, aliases []string) error {
	ws, err := s.workspaceRepo.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return fmt.Errorf("workspace %s not found: %w", workspaceSlug, err)
	}
	ws.Aliases = aliases
	if err := s.workspaceRepo.Update(ctx, ws); err != nil {
		return fmt.Errorf("failed to update workspace aliases: %w", err)
	}
	s.cache.Invalidate(workspaceSlug + ":")
	return nil
}

func (s *MemoryService) DetectWorkspace(ctx context.Context, path string) (*WorkspaceDetectionResult, error) {
	workspaces, err := s.workspaceRepo.List(ctx)
	if err != nil {
		s.metrics.RecordDetection(false)
		return nil, err
	}

	path = strings.ReplaceAll(path, "\\", "/")
	components := strings.Split(path, "/")
	var cleanComponents []string
	for _, c := range components {
		c = strings.TrimSpace(c)
		if c != "" {
			cleanComponents = append(cleanComponents, strings.ToLower(c))
		}
	}

	if len(cleanComponents) == 0 {
		s.metrics.RecordDetection(false)
		return &WorkspaceDetectionResult{
			WorkspaceSlug: "",
			Score:         0,
			Confidence:    "none",
		}, nil
	}

	var bestWorkspace string
	var bestScore float64

	for _, ws := range workspaces {
		wsSlug := strings.ToLower(ws.Slug)
		var wsScore float64

		// Check primary slug
		for i, comp := range cleanComponents {
			if comp == wsSlug {
				distance := len(cleanComponents) - 1 - i
				score := 100.0 / float64(distance+1)
				if score > wsScore {
					wsScore = score
				}
			}
		}

		// Check aliases
		for _, alias := range ws.Aliases {
			aliasLower := strings.ToLower(alias)
			for i, comp := range cleanComponents {
				if comp == aliasLower {
					distance := len(cleanComponents) - 1 - i
					score := 80.0 / float64(distance+1)
					if score > wsScore {
						wsScore = score
					}
				}
			}
		}

		if wsScore > bestScore {
			bestScore = wsScore
			bestWorkspace = ws.Slug
		}
	}

	confidence := "none"
	if bestScore >= 50.0 {
		confidence = "high"
	} else if bestScore >= 20.0 {
		confidence = "medium"
	} else if bestScore > 0.0 {
		confidence = "low"
	}

	success := bestScore > 0.0
	s.metrics.RecordDetection(success)

	return &WorkspaceDetectionResult{
		WorkspaceSlug: bestWorkspace,
		Score:         bestScore,
		Confidence:    confidence,
	}, nil
}

func (s *MemoryService) recommendTagsForTask(ctx context.Context, task string) []string {
	if task == "" {
		return []string{}
	}
	prompt := fmt.Sprintf(`Given a task description, recommend a list of relevant technical tags (lowercase, kebab-case) that should be used to catalog memories related to this task.
Respond ONLY in JSON format matching this schema:
{
  "tags": [string]
}

Task description:
%s`, task)

	respText, err := s.llmSvc.Generate(ctx, prompt, true)
	if err != nil {
		return []string{}
	}

	var result struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(respText), &result); err != nil {
		return []string{}
	}
	return result.Tags
}

func (s *MemoryService) Bootstrap(ctx context.Context, workspaceSlug string, task string, maxMemories int) (*BootstrapResponse, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordBootstrap(time.Since(start))
	}()

	cacheKey := workspaceSlug + ":bootstrap:" + fmt.Sprintf("%s:%d", task, maxMemories)
	if val, ok := s.cache.Get(cacheKey); ok {
		if resp, ok := val.(*BootstrapResponse); ok {
			copyResp := *resp
			copyResp.Stats.CacheHit = true
			return &copyResp, nil
		}
	}

	ws, err := s.workspaceRepo.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return nil, fmt.Errorf("workspace %s not found: %w", workspaceSlug, err)
	}

	wsContext, err := s.WorkspaceContext(ctx, workspaceSlug)
	if err != nil {
		wsContext = ""
	}

	projSnapshot, err := s.ProjectSnapshot(ctx, workspaceSlug)
	if err != nil {
		projSnapshot = ""
	}

	allMemories, err := s.memoryItemRepo.ListByWorkspace(ctx, ws.ID, nil, 100)
	if err != nil {
		allMemories = []*domain.MemoryItem{}
	}

	var recentDecisions []*domain.MemoryItem
	var knownRisks []*domain.MemoryItem

	for _, item := range allMemories {
		isDecision := (item.MemoryType != nil && *item.MemoryType == "decision") || item.Type == domain.TypeDecision || hasTagString(item.Tags, "decision")
		if isDecision && len(recentDecisions) < 10 {
			recentDecisions = append(recentDecisions, item)
		}

		isRisk := (item.MemoryType != nil && *item.MemoryType == "risk") || hasTagString(item.Tags, "risk") || hasTagString(item.Tags, "danger") || hasTagString(item.Tags, "warning")
		if isRisk && len(knownRisks) < 10 {
			knownRisks = append(knownRisks, item)
		}
	}

	var relevantMemories []*domain.MemorySearchRow
	if task != "" {
		limit := maxMemories
		if limit <= 0 {
			limit = 5
		}
		relevantMemories, _ = s.Search(ctx, &SearchMemoryInput{
			WorkspaceSlug: workspaceSlug,
			Query:         task,
			Limit:         &limit,
		})
	}

	recommendedTags := s.recommendTagsForTask(ctx, task)

	projects, _ := s.projectRepo.ListByWorkspace(ctx, ws.ID)
	totalProjects := len(projects)

	totalMemories, _ := s.memoryItemRepo.CountByWorkspace(ctx, ws.ID)

	wsName := ws.Slug
	if ws.Name != nil && *ws.Name != "" {
		wsName = *ws.Name
	}

	resp := &BootstrapResponse{
		WorkspaceSlug:    ws.Slug,
		WorkspaceName:    wsName,
		WorkspaceContext: wsContext,
		ProjectSnapshot:  projSnapshot,
		RecentDecisions:  recentDecisions,
		KnownRisks:       knownRisks,
		RelevantMemories: relevantMemories,
		RecommendedTags:  recommendedTags,
		Stats: BootstrapStats{
			TotalMemories: totalMemories,
			TotalProjects: totalProjects,
			CacheHit:      false,
		},
	}

	s.cache.Set(cacheKey, resp, 5*time.Minute)

	return resp, nil
}

func (s *MemoryService) SmartBootstrap(ctx context.Context, path string, task string, maxMemories int) (*BootstrapResponse, error) {
	detectRes, err := s.DetectWorkspace(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("smart bootstrap failed workspace detection: %w", err)
	}
	if detectRes.WorkspaceSlug == "" {
		return nil, fmt.Errorf("no workspace detected for path: %s", path)
	}
	return s.Bootstrap(ctx, detectRes.WorkspaceSlug, task, maxMemories)
}

func hasTagString(tags []string, target string) bool {
	target = strings.ToLower(target)
	for _, t := range tags {
		if strings.ToLower(t) == target {
			return true
		}
	}
	return false
}

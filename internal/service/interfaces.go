package service

import (
	"context"

	"memory-brain/internal/domain"
)

type WorkspaceRepository interface {
	GetBySlug(ctx context.Context, slug string) (*domain.Workspace, error)
	Create(ctx context.Context, ws *domain.Workspace) error
	List(ctx context.Context) ([]*domain.Workspace, error)
	Count(ctx context.Context) (int, error)
	Update(ctx context.Context, ws *domain.Workspace) error
}

type ProjectRepository interface {
	GetBySlug(ctx context.Context, workspaceID string, slug string) (*domain.Project, error)
	Create(ctx context.Context, project *domain.Project) error
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Project, error)
}

type MemoryItemRepository interface {
	Create(ctx context.Context, item *domain.MemoryItem) error
	FindByID(ctx context.Context, id string) (*domain.MemoryItem, error)
	Update(ctx context.Context, item *domain.MemoryItem) error
	Search(ctx context.Context, workspaceID string, projectID *string, queryEmbedding []float32, types []domain.MemoryType, tags []string, onlyActive bool, limit int) ([]*domain.MemorySearchRow, error)
	ListByWorkspace(ctx context.Context, workspaceID string, types []domain.MemoryType, limit int) ([]*domain.MemoryItem, error)
	GetStats(ctx context.Context) (totalMemories int, vectorIndexSize int, err error)
	CountByWorkspace(ctx context.Context, workspaceID string) (int, error)
}

type MemoryLinkRepository interface {
	Create(ctx context.Context, link *domain.MemoryLink) error
	GetLinksForMemory(ctx context.Context, memoryID string) ([]*domain.MemoryLink, error)
}

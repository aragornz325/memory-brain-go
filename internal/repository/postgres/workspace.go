package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"memory-brain/internal/domain"
)

type WorkspaceRepository struct {
	pool *pgxpool.Pool
}

func NewWorkspaceRepository(pool *pgxpool.Pool) *WorkspaceRepository {
	return &WorkspaceRepository{pool: pool}
}

func (r *WorkspaceRepository) GetBySlug(ctx context.Context, slug string) (*domain.Workspace, error) {
	query := `
		SELECT id, created_at, updated_at, slug, name 
		FROM memory.workspaces 
		WHERE slug = $1`

	var ws domain.Workspace
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&ws.ID,
		&ws.CreatedAt,
		&ws.UpdatedAt,
		&ws.Slug,
		&ws.Name,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to query workspace by slug: %w", err)
	}

	return &ws, nil
}

func (r *WorkspaceRepository) Create(ctx context.Context, ws *domain.Workspace) error {
	var query string
	var err error

	if ws.ID != "" {
		query = `
			INSERT INTO memory.workspaces (id, slug, name, created_at, updated_at) 
			VALUES ($1, $2, $3, COALESCE($4, now()), COALESCE($5, now())) 
			RETURNING id, created_at, updated_at`
		err = r.pool.QueryRow(ctx, query, ws.ID, ws.Slug, ws.Name, ws.CreatedAt, ws.UpdatedAt).Scan(&ws.ID, &ws.CreatedAt, &ws.UpdatedAt)
	} else {
		query = `
			INSERT INTO memory.workspaces (slug, name) 
			VALUES ($1, $2) 
			RETURNING id, created_at, updated_at`
		err = r.pool.QueryRow(ctx, query, ws.Slug, ws.Name).Scan(&ws.ID, &ws.CreatedAt, &ws.UpdatedAt)
	}

	if err != nil {
		return fmt.Errorf("failed to insert workspace: %w", err)
	}

	return nil
}

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"memory-brain/internal/domain"
)

type ProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, workspaceID string, slug string) (*domain.Project, error) {
	query := `
		SELECT id, created_at, updated_at, workspace_id, slug 
		FROM memory.projects 
		WHERE workspace_id = $1 AND slug = $2`

	var p domain.Project
	err := r.pool.QueryRow(ctx, query, workspaceID, slug).Scan(
		&p.ID,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.WorkspaceID,
		&p.Slug,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to query project by slug: %w", err)
	}

	return &p, nil
}

func (r *ProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	var query string
	var err error

	if p.ID != "" {
		query = `
			INSERT INTO memory.projects (id, workspace_id, slug, created_at, updated_at) 
			VALUES ($1, $2, $3, COALESCE($4, now()), COALESCE($5, now())) 
			RETURNING id, created_at, updated_at`
		err = r.pool.QueryRow(ctx, query, p.ID, p.WorkspaceID, p.Slug, p.CreatedAt, p.UpdatedAt).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	} else {
		query = `
			INSERT INTO memory.projects (workspace_id, slug) 
			VALUES ($1, $2) 
			RETURNING id, created_at, updated_at`
		err = r.pool.QueryRow(ctx, query, p.WorkspaceID, p.Slug).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	}

	if err != nil {
		return fmt.Errorf("failed to insert project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
	query := `
		SELECT id, created_at, updated_at, workspace_id, slug 
		FROM memory.projects 
		WHERE workspace_id = $1 
		ORDER BY slug ASC`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects by workspace: %w", err)
	}
	defer rows.Close()

	var list []*domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt, &p.WorkspaceID, &p.Slug); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		list = append(list, &p)
	}
	return list, nil
}

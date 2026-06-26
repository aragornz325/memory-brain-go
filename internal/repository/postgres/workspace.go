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
		SELECT id, created_at, updated_at, slug, name, aliases 
		FROM memory.workspaces 
		WHERE slug = $1`

	var ws domain.Workspace
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&ws.ID,
		&ws.CreatedAt,
		&ws.UpdatedAt,
		&ws.Slug,
		&ws.Name,
		&ws.Aliases,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to query workspace by slug: %w", err)
	}

	if ws.Aliases == nil {
		ws.Aliases = []string{}
	}

	return &ws, nil
}

func (r *WorkspaceRepository) Create(ctx context.Context, ws *domain.Workspace) error {
	var query string
	var err error

	aliases := ws.Aliases
	if aliases == nil {
		aliases = []string{}
	}

	if ws.ID != "" {
		query = `
			INSERT INTO memory.workspaces (id, slug, name, aliases, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, COALESCE($5, now()), COALESCE($6, now())) 
			RETURNING id, created_at, updated_at`
		err = r.pool.QueryRow(ctx, query, ws.ID, ws.Slug, ws.Name, aliases, ws.CreatedAt, ws.UpdatedAt).Scan(&ws.ID, &ws.CreatedAt, &ws.UpdatedAt)
	} else {
		query = `
			INSERT INTO memory.workspaces (slug, name, aliases) 
			VALUES ($1, $2, $3) 
			RETURNING id, created_at, updated_at`
		err = r.pool.QueryRow(ctx, query, ws.Slug, ws.Name, aliases).Scan(&ws.ID, &ws.CreatedAt, &ws.UpdatedAt)
	}

	if err != nil {
		return fmt.Errorf("failed to insert workspace: %w", err)
	}

	ws.Aliases = aliases
	return nil
}

func (r *WorkspaceRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	query := `
		SELECT id, created_at, updated_at, slug, name, aliases 
		FROM memory.workspaces 
		ORDER BY slug ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var list []*domain.Workspace
	for rows.Next() {
		var ws domain.Workspace
		if err := rows.Scan(&ws.ID, &ws.CreatedAt, &ws.UpdatedAt, &ws.Slug, &ws.Name, &ws.Aliases); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		if ws.Aliases == nil {
			ws.Aliases = []string{}
		}
		list = append(list, &ws)
	}
	return list, nil
}

func (r *WorkspaceRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM memory.workspaces`
	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count workspaces: %w", err)
	}
	return count, nil
}

func (r *WorkspaceRepository) Update(ctx context.Context, ws *domain.Workspace) error {
	aliases := ws.Aliases
	if aliases == nil {
		aliases = []string{}
	}

	query := `
		UPDATE memory.workspaces 
		SET name = $1, aliases = $2, updated_at = now() 
		WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, ws.Name, aliases, ws.ID)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	ws.Aliases = aliases
	return nil
}

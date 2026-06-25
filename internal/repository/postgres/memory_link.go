package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"memory-brain/internal/domain"
)

type MemoryLinkRepository struct {
	pool *pgxpool.Pool
}

func NewMemoryLinkRepository(pool *pgxpool.Pool) *MemoryLinkRepository {
	return &MemoryLinkRepository{pool: pool}
}

func (r *MemoryLinkRepository) Create(ctx context.Context, link *domain.MemoryLink) error {
	query := `
		INSERT INTO memory.memory_links (from_memory_id, to_memory_id, relation_type) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, link.FromMemoryID, link.ToMemoryID, link.RelationType).Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert memory link: %w", err)
	}

	return nil
}

func (r *MemoryLinkRepository) GetLinksForMemory(ctx context.Context, memoryID string) ([]*domain.MemoryLink, error) {
	query := `
		SELECT id, created_at, updated_at, from_memory_id, to_memory_id, relation_type 
		FROM memory.memory_links 
		WHERE from_memory_id = $1 OR to_memory_id = $1`

	rows, err := r.pool.Query(ctx, query, memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory links: %w", err)
	}
	defer rows.Close()

	var links []*domain.MemoryLink
	for rows.Next() {
		var link domain.MemoryLink
		err := rows.Scan(
			&link.ID,
			&link.CreatedAt,
			&link.UpdatedAt,
			&link.FromMemoryID,
			&link.ToMemoryID,
			&link.RelationType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory link row: %w", err)
		}
		links = append(links, &link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading memory link rows: %w", err)
	}

	return links, nil
}

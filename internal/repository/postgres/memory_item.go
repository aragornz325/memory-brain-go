package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"memory-brain/internal/domain"
)

type MemoryItemRepository struct {
	pool *pgxpool.Pool
}

func NewMemoryItemRepository(pool *pgxpool.Pool) *MemoryItemRepository {
	return &MemoryItemRepository{pool: pool}
}

func (r *MemoryItemRepository) Create(ctx context.Context, item *domain.MemoryItem) error {
	query := `
		INSERT INTO memory.memory_items (
			workspace_id, project_id, type, title, content, summary, tags, source, source_ref, metadata, importance, confidence, is_active, embedding
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14::vector)
		RETURNING id, created_at, updated_at`

	metadataBytes, err := json.Marshal(item.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var embeddingParam *string
	if len(item.Embedding) > 0 {
		str := float32SliceToVectorString(item.Embedding)
		embeddingParam = &str
	}

	err = r.pool.QueryRow(ctx, query,
		item.WorkspaceID,
		item.ProjectID,
		string(item.Type),
		item.Title,
		item.Content,
		item.Summary,
		item.Tags,
		item.Source,
		item.SourceRef,
		metadataBytes,
		item.Importance,
		item.Confidence,
		item.IsActive,
		embeddingParam,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert memory item: %w", err)
	}

	return nil
}

func (r *MemoryItemRepository) FindByID(ctx context.Context, id string) (*domain.MemoryItem, error) {
	query := `
		SELECT 
			mi.id, mi.created_at, mi.updated_at, mi.workspace_id, mi.project_id, 
			mi.type, mi.title, mi.content, mi.summary, mi.tags, 
			mi.source, mi.source_ref, mi.metadata, mi.importance, mi.confidence, 
			mi.is_active, w.slug AS workspace_slug, p.slug AS project_slug
		FROM memory.memory_items mi
		JOIN memory.workspaces w ON mi.workspace_id = w.id
		LEFT JOIN memory.projects p ON mi.project_id = p.id
		WHERE mi.id = $1`

	var item domain.MemoryItem
	var metadataBytes []byte
	var itemType string
	var projectSlugNullable *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.WorkspaceID,
		&item.ProjectID,
		&itemType,
		&item.Title,
		&item.Content,
		&item.Summary,
		&item.Tags,
		&item.Source,
		&item.SourceRef,
		&metadataBytes,
		&item.Importance,
		&item.Confidence,
		&item.IsActive,
		&item.WorkspaceSlug,
		&projectSlugNullable,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMemoryNotFound
		}
		return nil, fmt.Errorf("failed to find memory item: %w", err)
	}

	item.Type = domain.MemoryType(itemType)
	item.ProjectSlug = projectSlugNullable

	if len(metadataBytes) > 0 {
		if err := json.Unmarshal(metadataBytes, &item.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		item.Metadata = make(map[string]interface{})
	}

	return &item, nil
}

func (r *MemoryItemRepository) Update(ctx context.Context, item *domain.MemoryItem) error {
	query := `
		UPDATE memory.memory_items
		SET type = $1, title = $2, content = $3, summary = $4, tags = $5, source = $6, source_ref = $7, metadata = $8, importance = $9, confidence = $10, is_active = $11, embedding = COALESCE($12::vector, embedding), updated_at = now()
		WHERE id = $13
		RETURNING updated_at`

	metadataBytes, err := json.Marshal(item.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var embeddingParam *string
	if len(item.Embedding) > 0 {
		str := float32SliceToVectorString(item.Embedding)
		embeddingParam = &str
	}

	err = r.pool.QueryRow(ctx, query,
		string(item.Type),
		item.Title,
		item.Content,
		item.Summary,
		item.Tags,
		item.Source,
		item.SourceRef,
		metadataBytes,
		item.Importance,
		item.Confidence,
		item.IsActive,
		embeddingParam,
		item.ID,
	).Scan(&item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update memory item: %w", err)
	}

	return nil
}

func (r *MemoryItemRepository) Search(
	ctx context.Context,
	workspaceID string,
	projectID *string,
	queryEmbedding []float32,
	types []domain.MemoryType,
	tags []string,
	onlyActive bool,
	limit int,
) ([]*domain.MemorySearchRow, error) {
	embeddingStr := float32SliceToVectorString(queryEmbedding)

	params := []interface{}{
		workspaceID,
		projectID,
		onlyActive,
		embeddingStr,
		limit,
	}

	typeFilterSql := ""
	if len(types) > 0 {
		typeStrings := make([]string, len(types))
		for i, t := range types {
			typeStrings[i] = string(t)
		}
		params = append(params, typeStrings)
		typeFilterSql = fmt.Sprintf("AND mi.type = ANY($%d::text[])", len(params))
	}

	tagsFilterSql := ""
	if len(tags) > 0 {
		params = append(params, tags)
		tagsFilterSql = fmt.Sprintf("AND mi.tags && $%d::text[]", len(params))
	}

	sql := fmt.Sprintf(`
		SELECT
			mi.id,
			mi.workspace_id,
			mi.project_id,
			mi.type,
			mi.title,
			mi.content,
			mi.summary,
			mi.tags,
			mi.source,
			mi.source_ref,
			mi.metadata,
			mi.importance,
			mi.confidence,
			mi.is_active,
			1 - (mi.embedding <=> $4::vector) AS score,
			mi.created_at,
			mi.updated_at,
			w.slug AS workspace_slug,
			p.slug AS project_slug
		FROM memory.memory_items mi
		JOIN memory.workspaces w ON mi.workspace_id = w.id
		LEFT JOIN memory.projects p ON mi.project_id = p.id
		WHERE mi.workspace_id = $1::uuid
			AND ($2::uuid IS NULL OR mi.project_id = $2::uuid)
			AND ($3::boolean = false OR mi.is_active = true)
			AND mi.embedding IS NOT NULL
			%s
			%s
		ORDER BY mi.embedding <=> $4::vector
		LIMIT $5::int`,
		typeFilterSql,
		tagsFilterSql,
	)

	rows, err := r.pool.Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}
	defer rows.Close()

	var results []*domain.MemorySearchRow
	for rows.Next() {
		var row domain.MemorySearchRow
		var itemType string
		var metadataBytes []byte
		var projectSlugNullable *string

		err := rows.Scan(
			&row.ID,
			&row.WorkspaceID,
			&row.ProjectID,
			&itemType,
			&row.Title,
			&row.Content,
			&row.Summary,
			&row.Tags,
			&row.Source,
			&row.SourceRef,
			&metadataBytes,
			&row.Importance,
			&row.Confidence,
			&row.IsActive,
			&row.Score,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.WorkspaceSlug,
			&projectSlugNullable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search row: %w", err)
		}

		row.Type = domain.MemoryType(itemType)
		row.ProjectSlug = projectSlugNullable

		if len(metadataBytes) > 0 {
			if err := json.Unmarshal(metadataBytes, &row.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata in search: %w", err)
			}
		} else {
			row.Metadata = make(map[string]interface{})
		}

		results = append(results, &row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading search result rows: %w", err)
	}

	return results, nil
}

func float32SliceToVectorString(slice []float32) string {
	if len(slice) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i, f := range slice {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(fmt.Sprintf("%g", f))
	}
	sb.WriteByte(']')
	return sb.String()
}

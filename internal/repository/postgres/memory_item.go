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
	status := item.Status
	if status == "" {
		status = domain.StatusActive
	}

	query := `
		INSERT INTO memory.memory_items (
			workspace_id, project_id, type, title, content, summary, tags, source, source_ref, metadata, importance, confidence, is_active, embedding, memory_type, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14::vector, $15, $16)
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
		item.MemoryType,
		status,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert memory item: %w", err)
	}

	item.Status = status
	return nil
}

func (r *MemoryItemRepository) FindByID(ctx context.Context, id string) (*domain.MemoryItem, error) {
	query := `
		SELECT 
			mi.id, mi.created_at, mi.updated_at, mi.workspace_id, mi.project_id, 
			mi.type, mi.title, mi.content, mi.summary, mi.tags, 
			mi.source, mi.source_ref, mi.metadata, mi.importance, mi.confidence, 
			mi.is_active, w.slug AS workspace_slug, p.slug AS project_slug,
			mi.memory_type, mi.status
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
		&item.MemoryType,
		&item.Status,
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
	status := item.Status
	if status == "" {
		status = domain.StatusActive
	}

	query := `
		UPDATE memory.memory_items
		SET type = $1, title = $2, content = $3, summary = $4, tags = $5, source = $6, source_ref = $7, metadata = $8, importance = $9, confidence = $10, is_active = $11, embedding = COALESCE($12::vector, embedding), memory_type = $13, status = $14, updated_at = now()
		WHERE id = $15
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
		item.MemoryType,
		status,
		item.ID,
	).Scan(&item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update memory item: %w", err)
	}

	item.Status = status
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
			1 - (mi.embedding <=> $4::vector) + (mi.importance::float / 10.0 * 0.15) AS score,
			mi.created_at,
			mi.updated_at,
			w.slug AS workspace_slug,
			p.slug AS project_slug,
			mi.memory_type,
			mi.status
		FROM memory.memory_items mi
		JOIN memory.workspaces w ON mi.workspace_id = w.id
		LEFT JOIN memory.projects p ON mi.project_id = p.id
		WHERE mi.workspace_id = $1::uuid
			AND ($2::uuid IS NULL OR mi.project_id = $2::uuid)
			AND ($3::boolean = false OR (mi.is_active = true AND mi.status = 'active'))
			AND mi.embedding IS NOT NULL
			%s
			%s
		ORDER BY (mi.embedding <=> $4::vector) - (mi.importance::float / 10.0 * 0.15) ASC
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
			&row.MemoryType,
			&row.Status,
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

func (r *MemoryItemRepository) ListByWorkspace(
	ctx context.Context,
	workspaceID string,
	types []domain.MemoryType,
	limit int,
) ([]*domain.MemoryItem, error) {
	params := []interface{}{workspaceID, limit}
	typeFilterSql := ""
	if len(types) > 0 {
		typeStrings := make([]string, len(types))
		for i, t := range types {
			typeStrings[i] = string(t)
		}
		params = append(params, typeStrings)
		typeFilterSql = fmt.Sprintf("AND mi.type = ANY($%d::text[])", len(params))
	}

	sql := fmt.Sprintf(`
		SELECT 
			mi.id, mi.created_at, mi.updated_at, mi.workspace_id, mi.project_id, 
			mi.type, mi.title, mi.content, mi.summary, mi.tags, 
			mi.source, mi.source_ref, mi.metadata, mi.importance, mi.confidence, 
			mi.is_active, w.slug AS workspace_slug, p.slug AS project_slug,
			mi.memory_type, mi.status
		FROM memory.memory_items mi
		JOIN memory.workspaces w ON mi.workspace_id = w.id
		LEFT JOIN memory.projects p ON mi.project_id = p.id
		WHERE mi.workspace_id = $1::uuid AND mi.is_active = true AND mi.status = 'active'
			%s
		ORDER BY mi.importance DESC, mi.created_at DESC
		LIMIT $2::int`,
		typeFilterSql,
	)

	rows, err := r.pool.Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories by workspace: %w", err)
	}
	defer rows.Close()

	var results []*domain.MemoryItem
	for rows.Next() {
		var item domain.MemoryItem
		var itemType string
		var metadataBytes []byte
		var projectSlugNullable *string

		err := rows.Scan(
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
			&item.MemoryType,
			&item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory item row: %w", err)
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

		results = append(results, &item)
	}

	return results, nil
}

func (r *MemoryItemRepository) GetStats(ctx context.Context) (totalMemories int, vectorIndexSize int, err error) {
	query := `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE embedding IS NOT NULL)
		FROM memory.memory_items 
		WHERE is_active = true`
	err = r.pool.QueryRow(ctx, query).Scan(&totalMemories, &vectorIndexSize)
	if err != nil {
		err = fmt.Errorf("failed to get memory stats: %w", err)
	}
	return
}

func (r *MemoryItemRepository) CountByWorkspace(ctx context.Context, workspaceID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM memory.memory_items 
		WHERE workspace_id = $1 AND is_active = true AND status = 'active'`
	var count int
	err := r.pool.QueryRow(ctx, query, workspaceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count memories by workspace: %w", err)
	}
	return count, nil
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

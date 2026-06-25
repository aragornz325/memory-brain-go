-- +goose Up
-- 1. Asegurar que existe el proyecto 'memory-brain' en el workspace 'personal-lab'
INSERT INTO memory.projects (workspace_id, slug)
SELECT id, 'memory-brain'
FROM memory.workspaces
WHERE slug = 'personal-lab'
ON CONFLICT (workspace_id, slug) DO NOTHING;

-- 2. Migrar las memorias correspondientes desde el proyecto 'titan' al proyecto 'memory-brain'
UPDATE memory.memory_items
SET project_id = (
    SELECT p.id 
    FROM memory.projects p 
    JOIN memory.workspaces w ON p.workspace_id = w.id 
    WHERE w.slug = 'personal-lab' AND p.slug = 'memory-brain'
)
WHERE workspace_id = (
    SELECT id FROM memory.workspaces WHERE slug = 'personal-lab'
)
AND project_id = (
    SELECT p.id 
    FROM memory.projects p 
    JOIN memory.workspaces w ON p.workspace_id = w.id 
    WHERE w.slug = 'personal-lab' AND p.slug = 'titan'
)
AND (
    tags && ARRAY['go', 'chi', 'pgx', 'goose', 'slog', 'cobra', 'clean-architecture', 'architecture', 'logging', 'config', 'security', 'errors', 'memorybrain']
    OR title ILIKE '%Go%' OR content ILIKE '%Go%'
    OR title ILIKE '%architecture%' OR content ILIKE '%architecture%'
    OR title ILIKE '%Chi%' OR content ILIKE '%Chi%'
    OR title ILIKE '%pgx%' OR content ILIKE '%pgx%'
    OR title ILIKE '%slog%' OR content ILIKE '%slog%'
    OR title ILIKE '%Cobra%' OR content ILIKE '%Cobra%'
    OR title ILIKE '%Goose%' OR content ILIKE '%Goose%'
);

-- +goose Down
-- Revertir la asignación de proyecto a las memorias volviendo a asociarlas a 'titan'
UPDATE memory.memory_items
SET project_id = (
    SELECT p.id 
    FROM memory.projects p 
    JOIN memory.workspaces w ON p.workspace_id = w.id 
    WHERE w.slug = 'personal-lab' AND p.slug = 'titan'
)
WHERE workspace_id = (
    SELECT id FROM memory.workspaces WHERE slug = 'personal-lab'
)
AND project_id = (
    SELECT p.id 
    FROM memory.projects p 
    JOIN memory.workspaces w ON p.workspace_id = w.id 
    WHERE w.slug = 'personal-lab' AND p.slug = 'memory-brain'
)
AND (
    tags && ARRAY['go', 'chi', 'pgx', 'goose', 'slog', 'cobra', 'clean-architecture', 'architecture', 'logging', 'config', 'security', 'errors', 'memorybrain']
);

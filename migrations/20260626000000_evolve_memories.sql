-- +goose Up
ALTER TABLE memory.memory_items ADD COLUMN memory_type character varying;
ALTER TABLE memory.memory_items ADD COLUMN status character varying NOT NULL DEFAULT 'active';

-- +goose Down
ALTER TABLE memory.memory_items DROP COLUMN IF EXISTS memory_type;
ALTER TABLE memory.memory_items DROP COLUMN IF EXISTS status;

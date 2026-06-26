-- +goose Up
ALTER TABLE memory.workspaces ADD COLUMN aliases text[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE memory.workspaces DROP COLUMN IF EXISTS aliases;

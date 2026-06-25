-- +goose Up
-- Enable extensions if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- Create memory schema
CREATE SCHEMA IF NOT EXISTS "memory";

-- Workspaces table
CREATE TABLE "memory"."workspaces" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "slug" character varying NOT NULL,
    "name" character varying,
    CONSTRAINT "UQ_workspaces_slug" UNIQUE ("slug"),
    CONSTRAINT "PK_workspaces" PRIMARY KEY ("id")
);

-- Projects table
CREATE TABLE "memory"."projects" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "workspace_id" uuid NOT NULL,
    "slug" character varying NOT NULL,
    CONSTRAINT "uq_projects_workspace_slug" UNIQUE ("workspace_id", "slug"),
    CONSTRAINT "PK_projects" PRIMARY KEY ("id"),
    CONSTRAINT "FK_projects_workspace" FOREIGN KEY ("workspace_id") REFERENCES "memory"."workspaces"("id") ON DELETE CASCADE ON UPDATE NO ACTION
);

-- Memory items table
CREATE TABLE "memory"."memory_items" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "workspace_id" uuid NOT NULL,
    "project_id" uuid,
    "type" character varying NOT NULL,
    "title" character varying NOT NULL,
    "content" text NOT NULL,
    "summary" text,
    "tags" text array NOT NULL DEFAULT '{}',
    "source" character varying,
    "source_ref" character varying,
    "metadata" jsonb NOT NULL DEFAULT '{}',
    "importance" smallint NOT NULL DEFAULT '0',
    "confidence" numeric(3,2) NOT NULL DEFAULT '1',
    "is_active" boolean NOT NULL DEFAULT true,
    "embedding" vector,
    CONSTRAINT "PK_memory_items" PRIMARY KEY ("id"),
    CONSTRAINT "FK_memory_items_workspace" FOREIGN KEY ("workspace_id") REFERENCES "memory"."workspaces"("id") ON DELETE CASCADE ON UPDATE NO ACTION,
    CONSTRAINT "FK_memory_items_project" FOREIGN KEY ("project_id") REFERENCES "memory"."projects"("id") ON DELETE SET NULL ON UPDATE NO ACTION
);

-- Memory links table
CREATE TABLE "memory"."memory_links" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "from_memory_id" uuid NOT NULL,
    "to_memory_id" uuid NOT NULL,
    "relation_type" character varying NOT NULL,
    CONSTRAINT "uq_memory_links_relation" UNIQUE ("from_memory_id", "to_memory_id", "relation_type"),
    CONSTRAINT "PK_memory_links" PRIMARY KEY ("id"),
    CONSTRAINT "FK_memory_links_from" FOREIGN KEY ("from_memory_id") REFERENCES "memory"."memory_items"("id") ON DELETE CASCADE ON UPDATE NO ACTION,
    CONSTRAINT "FK_memory_links_to" FOREIGN KEY ("to_memory_id") REFERENCES "memory"."memory_items"("id") ON DELETE CASCADE ON UPDATE NO ACTION
);

-- +goose Down
DROP TABLE IF EXISTS "memory"."memory_links" CASCADE;
DROP TABLE IF EXISTS "memory"."memory_items" CASCADE;
DROP TABLE IF EXISTS "memory"."projects" CASCADE;
DROP TABLE IF EXISTS "memory"."workspaces" CASCADE;
DROP SCHEMA IF EXISTS "memory" CASCADE;

package domain

import "errors"

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrProjectNotFound   = errors.New("project not found")
	ErrMemoryNotFound    = errors.New("memory item not found")
	ErrInvalidEmbedding  = errors.New("failed to generate embedding")
	ErrDuplicateMemory   = errors.New("memory item already exists")
	ErrUnauthorized      = errors.New("unauthorized: invalid or missing API Key")
)

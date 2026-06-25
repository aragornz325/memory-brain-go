package domain

import (
	"time"
)

type Project struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	WorkspaceID string    `json:"workspace_id"`
	Slug        string    `json:"slug"`
}

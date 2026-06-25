package domain

import (
	"time"
)

type Workspace struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Slug      string    `json:"slug"`
	Name      *string   `json:"name,omitempty"`
}

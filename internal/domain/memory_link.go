package domain

import (
	"time"
)

type MemoryLink struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	FromMemoryID string    `json:"from_memory_id"`
	ToMemoryID   string    `json:"to_memory_id"`
	RelationType string    `json:"relation_type"`
}

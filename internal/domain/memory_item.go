package domain

import (
	"time"
)

type MemoryType string

const (
	TypeProject         MemoryType = "PROJECT"
	TypeDecision        MemoryType = "DECISION"
	TypeCode            MemoryType = "CODE"
	TypeConversation    MemoryType = "CONVERSATION"
	TypeInfra           MemoryType = "INFRA"
	TypeBug             MemoryType = "BUG"
	TypeIncident        MemoryType = "INCIDENT"
	TypeTroubleshooting MemoryType = "TROUBLESHOOTING"
	TypeRunbook         MemoryType = "RUNBOOK"
	TypeFAQ             MemoryType = "FAQ"
	TypeKnownError      MemoryType = "KNOWN_ERROR"
	TypePrompt          MemoryType = "PROMPT"
)

type MemoryItem struct {
	ID            string                 `json:"id"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	WorkspaceID   string                 `json:"workspace_id"`
	ProjectID     *string                `json:"project_id"`
	WorkspaceSlug string                 `json:"workspace_slug,omitempty"`
	ProjectSlug   *string                `json:"project_slug,omitempty"`
	Type          MemoryType             `json:"type"`
	Title         string                 `json:"title"`
	Content       string                 `json:"content"`
	Summary       *string                `json:"summary"`
	Tags          []string               `json:"tags"`
	Source        *string                `json:"source"`
	SourceRef     *string                `json:"source_ref"`
	Metadata      map[string]interface{} `json:"metadata"`
	Importance    int                    `json:"importance"`
	Confidence    float64                `json:"confidence"`
	IsActive      bool                   `json:"is_active"`
	Embedding     []float32              `json:"-"` // Not selected by default
}

type MemorySearchRow struct {
	MemoryItem
	Score float64 `json:"score"`
}

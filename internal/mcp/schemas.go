package mcp

// SearchArgs defines the parameters for the memory.search tool.
type SearchArgs struct {
	Query         string `json:"query" jsonschema:"The search term or query to find relevant memories semantically"`
	WorkspaceSlug string `json:"workspaceSlug,omitempty" jsonschema:"Workspace slug to search within (optional, defaults to environment config)"`
	ProjectSlug   string `json:"projectSlug,omitempty" jsonschema:"Project slug to search within (optional, defaults to environment config)"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum number of search results to return (optional, defaults to 10)"`
}

// RememberArgs defines the parameters for the memory.remember tool.
type RememberArgs struct {
	Text          string   `json:"text" jsonschema:"The context, facts, decisions, or text content to store as a memory"`
	WorkspaceSlug string   `json:"workspaceSlug,omitempty" jsonschema:"Workspace slug to save in (optional, defaults to environment config)"`
	ProjectSlug   string   `json:"projectSlug,omitempty" jsonschema:"Project slug to save in (optional, defaults to environment config)"`
	Tags          []string `json:"tags,omitempty" jsonschema:"Optional tags to categorize this memory"`
	Source        string   `json:"source,omitempty" jsonschema:"Optional source of the memory (defaults to mcp)"`
	SourceRef     string   `json:"sourceRef,omitempty" jsonschema:"Optional reference or file path/line details where this memory originated"`
}

// HealthArgs defines the parameters for the memory.health tool, which has no arguments.
type HealthArgs struct{}

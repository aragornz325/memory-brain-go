package mcp

// SearchArgs defines the parameters for the memory.search tool.
type SearchArgs struct {
	Query         string   `json:"query" jsonschema:"The search term or query to find relevant memories semantically"`
	WorkspaceSlug string   `json:"workspaceSlug,omitempty" jsonschema:"Workspace slug to search within (optional, defaults to environment config)"`
	ProjectSlug   string   `json:"projectSlug,omitempty" jsonschema:"Project slug to search within (optional, defaults to environment config)"`
	Limit         int      `json:"limit,omitempty" jsonschema:"Maximum number of search results to return (optional, defaults to 10)"`
	Types         []string `json:"types,omitempty" jsonschema:"Filter results by memory types (e.g. architecture, decision, bugfix)"`
}

// RememberArgs defines the parameters for the memory.remember tool.
type RememberArgs struct {
	Text          string   `json:"text" jsonschema:"The context, facts, decisions, or text content to store as a memory"`
	WorkspaceSlug string   `json:"workspaceSlug,omitempty" jsonschema:"Workspace slug to save in (optional, defaults to environment config)"`
	ProjectSlug   string   `json:"projectSlug,omitempty" jsonschema:"Project slug to save in (optional, defaults to environment config)"`
	Tags          []string `json:"tags,omitempty" jsonschema:"Optional tags to categorize this memory"`
	Source        string   `json:"source,omitempty" jsonschema:"Optional source of the memory (defaults to mcp)"`
	SourceRef     string   `json:"sourceRef,omitempty" jsonschema:"Optional reference or file path/line details where this memory originated"`
	MemoryType    string   `json:"memoryType,omitempty" jsonschema:"The type of development memory (e.g. architecture, decision, bugfix, workflow, convention, infrastructure, lesson_learned, configuration, deployment)"`
	Status        string   `json:"status,omitempty" jsonschema:"The obsolescence status (active, deprecated, obsolete)"`
	Importance    int      `json:"importance,omitempty" jsonschema:"Importance score from 1 to 10"`
}

// HealthArgs defines the parameters for the memory.health tool.
type HealthArgs struct{}

// WorkspaceContextArgs defines the parameters for memory.workspace_context.
type WorkspaceContextArgs struct {
	Workspace string `json:"workspace" jsonschema:"The slug of the workspace to retrieve context for"`
}

// ProjectSnapshotArgs defines the parameters for memory.project_snapshot.
type ProjectSnapshotArgs struct {
	Workspace string `json:"workspace" jsonschema:"The slug of the workspace to retrieve snapshot for"`
}

// RelatedArgs defines the parameters for memory.related.
type RelatedArgs struct {
	MemoryID string `json:"memoryId" jsonschema:"The ID of the memory to fetch relations for"`
}

// DecideSaveArgs defines the parameters for memory.decide_save.
type DecideSaveArgs struct {
	Summary string `json:"summary" jsonschema:"The summary/content of the memory to evaluate for saving"`
}

// StatsArgs defines the parameters for memory.stats.
type StatsArgs struct{}

// LinkArgs defines the parameters for memory.link.
type LinkArgs struct {
	FromMemoryID string `json:"fromMemoryId" jsonschema:"The ID of the source memory"`
	ToMemoryID   string `json:"toMemoryId" jsonschema:"The ID of the target memory"`
	RelationType string `json:"relationType" jsonschema:"The type of relationship (e.g. related_to, supersedes)"`
}

// BootstrapArgs defines the parameters for memory.bootstrap.
type BootstrapArgs struct {
	Workspace   string `json:"workspace" jsonschema:"The slug of the workspace to bootstrap"`
	CurrentTask string `json:"currentTask,omitempty" jsonschema:"The description of the task currently being initialized (optional)"`
	MaxMemories int    `json:"maxMemories,omitempty" jsonschema:"Maximum number of semantically relevant memories to retrieve for the task (optional, defaults to 5)"`
}

// DetectWorkspaceArgs defines the parameters for memory.detect_workspace.
type DetectWorkspaceArgs struct {
	Path string `json:"path" jsonschema:"The absolute file or folder path to detect the workspace for"`
}

// SmartBootstrapArgs defines the parameters for memory.smart_bootstrap.
type SmartBootstrapArgs struct {
	Path        string `json:"path" jsonschema:"The absolute file or folder path of the active workspace"`
	CurrentTask string `json:"currentTask,omitempty" jsonschema:"The description of the task currently being initialized (optional)"`
	MaxMemories int    `json:"maxMemories,omitempty" jsonschema:"Maximum number of semantically relevant memories to retrieve for the task (optional, defaults to 5)"`
}

// SetWorkspaceAliasesArgs defines the parameters for memory.set_workspace_aliases.
type SetWorkspaceAliasesArgs struct {
	Workspace string   `json:"workspace" jsonschema:"The slug of the workspace to set aliases for"`
	Aliases   []string `json:"aliases" jsonschema:"The list of aliases to assign to this workspace"`
}


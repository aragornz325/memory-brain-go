package cli

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type CliConfig struct {
	BaseURL       string
	APIKey        string
	WorkspaceSlug string
	ProjectSlug   string
}

func LoadCliConfig(flagWorkspace, flagProject string) *CliConfig {
	// Attempt to load .env from current working directory
	_ = godotenv.Load()

	url := getEnv("MEMORY_BRAIN_URL", "http://localhost:3210")
	apiKey := getEnv("MEMORY_BRAIN_API_KEY", "")

	ws := flagWorkspace
	if ws == "" {
		ws = getEnv("MEMORY_BRAIN_WORKSPACE", "")
	}

	proj := flagProject
	if proj == "" {
		proj = getEnv("MEMORY_BRAIN_PROJECT", "")
	}

	return &CliConfig{
		BaseURL:       strings.TrimSuffix(url, "/"),
		APIKey:        apiKey,
		WorkspaceSlug: ws,
		ProjectSlug:   proj,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallback
}

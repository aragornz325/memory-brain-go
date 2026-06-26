package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Ollama   OllamaConfig
	Auth     AuthConfig
}

type HTTPConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Logging  bool
}

type OllamaConfig struct {
	BaseURL string
	Model   string
}

type AuthConfig struct {
	APIKey string
}

func Load() *Config {
	// Attempt to load .env file from working directory
	if err := godotenv.Load(); err != nil {
		// Fallback to the absolute path of the project .env file
		if errFallback := godotenv.Load("/home/chivo/memory-brain-go/.env"); errFallback != nil {
			slog.Warn("No .env file found or error reading it, relying on environmental variables")
		}
	}

	port := getEnv("SERVER_PORT", "3210")
	dbHost := getEnv("POSTGRES_HOST_MEMORY_BRAIN", "localhost")
	dbPortStr := getEnv("POSTGRES_PORT_MEMORY_BRAIN", "5432")
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		slog.Warn("Invalid POSTGRES_PORT_MEMORY_BRAIN, using default 5432", "val", dbPortStr)
		dbPort = 5432
	}
	dbUser := getEnv("POSTGRES_USER_MEMORY_BRAIN", "memory_brain")
	dbPassword := getEnv("POSTGRES_PASSWORD_MEMORY_BRAIN", "memory_brain")
	dbName := getEnv("POSTGRES_DB_MEMORY_BRAIN", "memory_brain")
	dbLoggingStr := getEnv("DB_LOGGING_MEMORY_BRAIN", "true")
	dbLogging := parseBool(dbLoggingStr, true)

	ollamaBaseURL := getEnv("OLLAMA_BASE_URL", "http://localhost:11434")
	ollamaModel := getEnv("OLLAMA_EMBEDDING_MODEL", "nomic-embed-text")
	apiKey := getEnv("MEMORY_BRAIN_API_KEY", "")

	return &Config{
		HTTP: HTTPConfig{
			Port: port,
		},
		Database: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Database: dbName,
			Logging:  dbLogging,
		},
		Ollama: OllamaConfig{
			BaseURL: ollamaBaseURL,
			Model:   ollamaModel,
		},
		Auth: AuthConfig{
			APIKey: apiKey,
		},
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func parseBool(value string, fallback bool) bool {
	val := strings.ToLower(strings.TrimSpace(value))
	if val == "" {
		return fallback
	}
	return val == "true" || val == "1" || val == "yes"
}

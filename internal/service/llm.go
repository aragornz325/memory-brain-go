package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"memory-brain/internal/config"
)

type LLMService struct {
	baseURL string
	model   string
	client  *http.Client
}

type OllamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format,omitempty"`
}

type OllamaGenerateResponse struct {
	Response string `json:"response"`
}

func NewLLMService(cfg *config.Config) *LLMService {
	model := os.Getenv("OLLAMA_LLM_MODEL")
	if model == "" {
		model = "qwen2.5-coder:3b"
	}
	return &LLMService{
		baseURL: strings.TrimSuffix(cfg.Ollama.BaseURL, "/"),
		model:   model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *LLMService) Generate(ctx context.Context, prompt string, jsonFormat bool) (string, error) {
	url := fmt.Sprintf("%s/api/generate", s.baseURL)

	reqPayload := OllamaGenerateRequest{
		Model:  s.model,
		Prompt: prompt,
		Stream: false,
	}
	if jsonFormat {
		reqPayload.Format = "json"
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal generate request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create generate request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama generate call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama generate returned non-200 status: %d", resp.StatusCode)
	}

	var respPayload OllamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return "", fmt.Errorf("failed to decode generate response: %w", err)
	}

	return respPayload.Response, nil
}

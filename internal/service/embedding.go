package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"memory-brain/internal/config"
)

type EmbeddingService struct {
	baseURL string
	model   string
	client  *http.Client
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaResponse struct {
	Embedding []float32 `json:"embedding"`
}

func NewEmbeddingService(cfg *config.Config) *EmbeddingService {
	return &EmbeddingService{
		baseURL: strings.TrimSuffix(cfg.Ollama.BaseURL, "/"),
		model:   cfg.Ollama.Model,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *EmbeddingService) Embed(ctx context.Context, text string) ([]float32, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/api/embeddings", s.baseURL)

	reqPayload := OllamaRequest{
		Model:  s.model,
		Prompt: text,
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request for ollama: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned non-200 status code: %d", resp.StatusCode)
	}

	var respPayload OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	if len(respPayload.Embedding) == 0 {
		return nil, errors.New("empty embedding vector returned from ollama")
	}

	return respPayload.Embedding, nil
}

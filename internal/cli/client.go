package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	httpApi "memory-brain/internal/http"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

type RememberRequest struct {
	WorkspaceSlug string   `json:"workspaceSlug"`
	ProjectSlug   *string  `json:"projectSlug,omitempty"`
	Text          string   `json:"text"`
	Tags          []string `json:"tags,omitempty"`
	Source        *string  `json:"source,omitempty"`
	SourceRef     *string  `json:"sourceRef,omitempty"`
}

type SearchRequest struct {
	WorkspaceSlug string   `json:"workspaceSlug"`
	ProjectSlug   *string  `json:"projectSlug,omitempty"`
	Query         string   `json:"query"`
	Limit         int      `json:"limit,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

type ContextResponse struct {
	Context  string                        `json:"context"`
	Memories []*httpApi.MemoryItemResponse `json:"memories"`
}

func (c *Client) Remember(ctx context.Context, req *RememberRequest) (*httpApi.MemoryItemResponse, error) {
	url := fmt.Sprintf("%s/memory/remember", c.baseURL)
	var resp httpApi.MemoryItemResponse
	if err := c.post(ctx, url, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Search(ctx context.Context, req *SearchRequest) ([]*httpApi.MemoryItemResponse, error) {
	url := fmt.Sprintf("%s/memory/search", c.baseURL)
	var resp []*httpApi.MemoryItemResponse
	if err := c.post(ctx, url, req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Context(ctx context.Context, req *SearchRequest) (*ContextResponse, error) {
	url := fmt.Sprintf("%s/memory/context", c.baseURL)
	var resp ContextResponse
	if err := c.post(ctx, url, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateWorkspaceRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type CreateProjectRequest struct {
	WorkspaceSlug string `json:"workspaceSlug"`
	Slug          string `json:"slug"`
}

type WorkspaceResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Slug      string    `json:"slug"`
	Name      *string   `json:"name"`
}

type ProjectResponse struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	WorkspaceID string    `json:"workspace_id"`
	Slug        string    `json:"slug"`
}

func (c *Client) CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*WorkspaceResponse, error) {
	url := fmt.Sprintf("%s/workspaces", c.baseURL)
	var resp WorkspaceResponse
	if err := c.post(ctx, url, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CreateProject(ctx context.Context, req *CreateProjectRequest) (*ProjectResponse, error) {
	url := fmt.Sprintf("%s/projects", c.baseURL)
	var resp ProjectResponse
	if err := c.post(ctx, url, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) post(ctx context.Context, url string, payload interface{}, target interface{}) error {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

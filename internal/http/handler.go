package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"memory-brain/internal/domain"
	"memory-brain/internal/service"
)

type Handler struct {
	memorySvc *service.MemoryService
}

func NewHandler(memorySvc *service.MemoryService) *Handler {
	return &Handler{memorySvc: memorySvc}
}

type errorResponse struct {
	Message string `json:"message"`
}

func (h *Handler) Hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello World!"))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateMemoryInput
	if !decodeJSON(w, r, &input) {
		return
	}

	if input.WorkspaceSlug == "" {
		respondWithError(w, http.StatusBadRequest, "workspaceSlug is required")
		return
	}
	if input.Type == "" {
		respondWithError(w, http.StatusBadRequest, "type is required")
		return
	}
	if input.Title == "" {
		respondWithError(w, http.StatusBadRequest, "title is required")
		return
	}
	if input.Content == "" {
		respondWithError(w, http.StatusBadRequest, "content is required")
		return
	}

	item, err := h.memorySvc.Create(r.Context(), &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, mapToResponse(item, input.WorkspaceSlug, input.ProjectSlug))
}

func (h *Handler) Remember(w http.ResponseWriter, r *http.Request) {
	var input service.RememberMemoryInput
	if !decodeJSON(w, r, &input) {
		return
	}

	if input.WorkspaceSlug == "" {
		respondWithError(w, http.StatusBadRequest, "workspaceSlug is required")
		return
	}
	if input.Text == "" {
		respondWithError(w, http.StatusBadRequest, "text is required")
		return
	}

	item, err := h.memorySvc.Remember(r.Context(), &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, mapToResponse(item, input.WorkspaceSlug, input.ProjectSlug))
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	var input service.SearchMemoryInput
	if !decodeJSON(w, r, &input) {
		return
	}

	if input.WorkspaceSlug == "" {
		respondWithError(w, http.StatusBadRequest, "workspaceSlug is required")
		return
	}
	if input.Query == "" {
		respondWithError(w, http.StatusBadRequest, "query is required")
		return
	}

	results, err := h.memorySvc.Search(r.Context(), &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	responseList := make([]*MemoryItemResponse, len(results))
	for i, row := range results {
		responseList[i] = mapSearchRowToResponse(row, input.WorkspaceSlug, input.ProjectSlug)
	}

	respondWithJSON(w, http.StatusOK, responseList)
}

func (h *Handler) Context(w http.ResponseWriter, r *http.Request) {
	var input service.SearchMemoryInput
	if !decodeJSON(w, r, &input) {
		return
	}

	if input.WorkspaceSlug == "" {
		respondWithError(w, http.StatusBadRequest, "workspaceSlug is required")
		return
	}
	if input.Query == "" {
		respondWithError(w, http.StatusBadRequest, "query is required")
		return
	}

	ctxResp, err := h.memorySvc.GetContext(r.Context(), &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	mappedMemories := make([]*MemoryItemResponse, len(ctxResp.Memories))
	for i, m := range ctxResp.Memories {
		mappedMemories[i] = mapSearchRowToResponse(m, input.WorkspaceSlug, input.ProjectSlug)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"context":  ctxResp.Context,
		"memories": mappedMemories,
	})
}

func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	item, err := h.memorySvc.FindOne(r.Context(), id)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, mapToResponse(item, item.WorkspaceSlug, item.ProjectSlug))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	var input service.UpdateMemoryInput
	if !decodeJSON(w, r, &input) {
		return
	}

	item, err := h.memorySvc.Update(r.Context(), id, &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, mapToResponse(item, item.WorkspaceSlug, item.ProjectSlug))
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "id parameter is required")
		return
	}

	err := h.memorySvc.Remove(r.Context(), id)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"success":true}`))
}

// Helpers
func decodeJSON(w http.ResponseWriter, r *http.Request, val interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(val); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respondWithError(w, http.StatusRequestEntityTooLarge, "Request payload too large (max 5MB)")
			return false
		}
		respondWithError(w, http.StatusBadRequest, "invalid JSON payload")
		return false
	}
	return true
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, errorResponse{Message: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"Internal Server Error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

func respondWithDomainError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrMemoryNotFound) {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, domain.ErrWorkspaceNotFound) {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, domain.ErrProjectNotFound) {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if errors.Is(err, domain.ErrDuplicateMemory) {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, domain.ErrInvalidEmbedding) {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	respondWithError(w, http.StatusInternalServerError, err.Error())
}

type CreateWorkspaceInput struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type CreateProjectInput struct {
	WorkspaceSlug string `json:"workspaceSlug"`
	Slug          string `json:"slug"`
}

func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var input CreateWorkspaceInput
	if !decodeJSON(w, r, &input) {
		return
	}

	if input.Slug == "" {
		respondWithError(w, http.StatusBadRequest, "slug is required")
		return
	}

	ws, err := h.memorySvc.CreateWorkspace(r.Context(), input.Slug, input.Name)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, ws)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var input CreateProjectInput
	if !decodeJSON(w, r, &input) {
		return
	}

	if input.WorkspaceSlug == "" {
		respondWithError(w, http.StatusBadRequest, "workspaceSlug is required")
		return
	}
	if input.Slug == "" {
		respondWithError(w, http.StatusBadRequest, "slug is required")
		return
	}

	p, err := h.memorySvc.CreateProject(r.Context(), input.WorkspaceSlug, input.Slug)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"memory-brain/internal/config"
)

func NewRouter(cfg *config.Config, h *Handler) *chi.Mux {
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(LoggerMiddleware)
	r.Use(MaxBytesMiddleware(5 * 1024 * 1024)) // Limit requests to 5MB
	r.Use(middleware.Recoverer)

	// Hello endpoint
	r.Get("/", h.Hello)

	// API endpoints with API key protection
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(cfg.Auth.APIKey))

		r.Post("/workspaces", h.CreateWorkspace)
		r.Post("/projects", h.CreateProject)

		r.Route("/memory", func(r chi.Router) {
			r.Post("/", h.Create)
			r.Post("/remember", h.Remember)
			r.Post("/search", h.Search)
			r.Post("/context", h.Context)
			r.Get("/{id}", h.FindOne)
			r.Patch("/{id}", h.Update)
			r.Delete("/{id}", h.Remove)
		})
	})

	return r
}

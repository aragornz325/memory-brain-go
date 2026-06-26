package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"memory-brain/internal/config"
	"memory-brain/internal/database"
	httpApi "memory-brain/internal/http"
	mcpServer "memory-brain/internal/mcp"
	"memory-brain/internal/repository/postgres"
	"memory-brain/internal/service"
)

func main() {
	// Configure slog to log JSON to stdout by default
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Check if we should boot as an MCP server
	isMcp := len(os.Args) > 1 && os.Args[1] == "mcp"

	if isMcp {
		// Override default logger to write to stderr so we don't pollute stdout
		// (Stdio transport protocol uses stdout for JSON-RPC messages)
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		slog.SetDefault(logger)
	}

	slog.Info("Starting Memory Brain Go backend...")

	// 1. Load config
	cfg := config.Load()

	// 2. Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 3. Initialize repositories
	workspaceRepo := postgres.NewWorkspaceRepository(db.Pool)
	projectRepo := postgres.NewProjectRepository(db.Pool)
	memoryItemRepo := postgres.NewMemoryItemRepository(db.Pool)

	// 4. Initialize services
	embeddingSvc := service.NewEmbeddingService(cfg)
	memorySvc := service.NewMemoryService(workspaceRepo, projectRepo, memoryItemRepo, embeddingSvc)

	if isMcp {
		slog.Info("Starting MCP Server mode...")
		mcpSrv := mcpServer.NewServer(memorySvc, db)
		if err := mcpSrv.Run(context.Background()); err != nil {
			slog.Error("MCP Server terminated with error", "error", err)
			os.Exit(1)
		}
		return
	}

	// 5. Initialize handlers and router
	handler := httpApi.NewHandler(memorySvc)
	router := httpApi.NewRouter(cfg, handler)

	// 6. Start server with graceful shutdown support
	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info(fmt.Sprintf("Memory Brain API is running on port %s", cfg.HTTP.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for termination signals to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful server shutdown failed", "error", err)
	} else {
		slog.Info("Server stopped gracefully")
	}
}

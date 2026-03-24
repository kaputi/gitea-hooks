package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kaputi/gitea-hooks/config"
	"github.com/kaputi/gitea-hooks/git"
	"github.com/kaputi/gitea-hooks/gitea"
	"github.com/kaputi/gitea-hooks/queue"
	"github.com/kaputi/gitea-hooks/reviewer"
	"github.com/kaputi/gitea-hooks/server"
	"github.com/kaputi/gitea-hooks/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Ensure clone base path exists
	if err := os.MkdirAll(cfg.CloneBasePath, 0755); err != nil {
		logger.Error("failed to create clone base path", "error", err)
		os.Exit(1)
	}

	// Initialize components
	q := queue.New(cfg.QueueSize)
	g := git.New(cfg.SSHKeyPath)
	r := reviewer.New(cfg.ClaudeSkill, cfg.AnthropicAPIKey)
	gc := gitea.New(cfg.GiteaURL, cfg.GiteaToken)
	w := worker.New(q, g, r, gc, cfg.CloneBasePath, cfg.RetentionHours, logger)
	srv := server.New(q, cfg.WebhookSecret, logger)

	// Start worker
	w.Start()

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: srv,
	}

	go func() {
		logger.Info("starting server", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}
	q.Close()
	w.Stop()
	logger.Info("shutdown complete")
}

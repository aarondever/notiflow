package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aarondever/notiflow/internal"
	"github.com/aarondever/notiflow/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	// Config logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize app with all dependencies via Wire
	app, err := internal.InitializeApp(cfg)
	if err != nil {
		slog.Error("Failed to initialize app", "error", err)
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_ = app.DB.Mongo.Disconnect(ctx)
	}()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	slog.Info("Starting web server", "address", addr)

	// Start server (blocking call)
	if err = app.Router.Run(addr); err != nil {
		slog.Error("Web server failed to start", "error", err, "address", addr)
		os.Exit(1)
	}
}

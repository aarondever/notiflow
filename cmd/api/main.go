package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
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

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start gRPC server in a goroutine
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			slog.Error("Failed to listen for gRPC", "error", err, "address", grpcAddr)
			os.Exit(1)
		}

		slog.Info("Starting gRPC server", "address", grpcAddr)
		if err := app.GRPCServer.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Start HTTP server in a goroutine
	httpAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	go func() {
		slog.Info("Starting HTTP server", "address", httpAddr)
		if err := app.Router.Run(httpAddr); err != nil {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	slog.Info("Shutting down servers...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop gRPC server
	app.GRPCServer.GracefulStop()

	// Disconnect from database
	if err := app.DB.Mongo.Disconnect(shutdownCtx); err != nil {
		slog.Error("Error disconnecting from database", "error", err)
	}

	slog.Info("Servers stopped gracefully")
}

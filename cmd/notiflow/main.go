package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aarondever/notiflow/internal/config"
	"github.com/aarondever/notiflow/internal/database"
	"github.com/aarondever/notiflow/internal/handlers"
	"github.com/aarondever/notiflow/internal/models"
	"github.com/aarondever/notiflow/internal/services"
	"github.com/aarondever/notiflow/internal/utils"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

type Application struct {
	webServer *http.Server
	metrics   *models.ApplicationMetrics
}

func main() {
	// Config logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Configuration loading failed", "error", err)
		os.Exit(1)
	}

	// Initialize database connection pool
	db, err := database.InitializeDatabase(cfg)
	if err != nil {
		slog.Error("Database initialization failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := db.Mongo.Disconnect(ctx); err != nil {
			slog.Error("Database disconnection error", "error", err)
		}
	}()

	// Initialize all services with dependency injection
	allServices := services.InitializeServices(db, cfg)

	// Initialize all handlers with service dependencies
	allHandlers := handlers.InitializeHandlers(allServices)

	// Configure middleware
	router := chi.NewRouter()
	router.Use(middleware.Logger)                    // Request logging
	router.Use(middleware.Recoverer)                 // Panic recovery
	router.Use(middleware.Compress(5))               // Response compression
	router.Use(middleware.Timeout(30 * time.Second)) // Request timeout

	// Setup routers
	allHandlers.SetupRouters(router)

	app := &Application{
		metrics: &models.ApplicationMetrics{StartTime: time.Now()},
	}

	// Configure server
	app.webServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup app routers
	router.Route("/api", func(r chi.Router) {
		router.Get("/health", app.getHealth)
		router.Get("/metrics", app.getMetrics)
	})

	// Setup graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		slog.Info("Received shutdown signal", "signal", sig)
		slog.Info("Initiating graceful shutdown...")

		app.initiateShutdown()
	}()

	slog.Info("Starting web server", "address", app.webServer.Addr)

	// Start server (blocking call)
	if err = app.webServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Web server failed to start", "error", err)
		os.Exit(1)
	}
}

// initiateShutdown begins the graceful shutdown process for all application components
func (app *Application) initiateShutdown() {
	shutdownStart := time.Now()

	// Stop web server with timeout
	if app.webServer != nil {
		slog.Info("Stopping web server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.webServer.Shutdown(ctx); err != nil {
			slog.Error("Web server shutdown error", "error", err)
		} else {
			slog.Info("Web server stopped gracefully")
		}
	}

	slog.Info("Graceful shutdown completed", "duration", time.Since(shutdownStart))
	slog.Info("Final application metrics",
		"start_time", app.metrics.StartTime.Format(time.RFC3339),
		"total_uptime", app.metrics.TotalUptime,
	)
}

func (app *Application) getHealth(w http.ResponseWriter, r *http.Request) {
	healthResponse := map[string]interface{}{
		"status":  "healthy",
		"service": "notiflow",
	}
	utils.RespondWithJSON(w, http.StatusOK, healthResponse)
}

func (app *Application) getMetrics(w http.ResponseWriter, r *http.Request) {
	app.metrics.TotalUptime = time.Since(app.metrics.StartTime)
	utils.RespondWithJSON(w, http.StatusOK, app.metrics)
}

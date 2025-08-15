package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aarondever/url-forg/internal/config"
	"github.com/aarondever/url-forg/internal/database"
	"github.com/aarondever/url-forg/internal/handlers"
	"github.com/aarondever/url-forg/internal/models"
	"github.com/aarondever/url-forg/internal/services"
	"github.com/go-chi/chi/v5/middleware"
	"log"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration loading failed: %v", err)
	}

	// Config timezone
	time.Local = cfg.Timezone
	log.Printf("Application timezone: %s\n", cfg.Timezone)

	// Initialize database connection pool
	db, err := database.InitializeDatabase(cfg)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := db.Mongo.Disconnect(ctx); err != nil {
			log.Printf("Database disconnection error: %v", err)
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

	// Setup graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Printf("Received shutdown signal: %v", sig)
		log.Printf("Initiating graceful shutdown...")

		app.initiateShutdown()
	}()

	log.Printf("Starting web server on %s", app.webServer.Addr)

	// Start server (blocking call)
	if err = app.webServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Web server failed to start: %v", err)
	}
}

// initiateShutdown begins the graceful shutdown process for all application components
func (app *Application) initiateShutdown() {
	shutdownStart := time.Now()

	// Stop web server with timeout
	if app.webServer != nil {
		log.Printf("Stopping web server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.webServer.Shutdown(ctx); err != nil {
			log.Printf("Web server shutdown error: %v", err)
		} else {
			log.Printf("Web server stopped gracefully")
		}
	}

	log.Printf("Graceful shutdown completed in %v", time.Since(shutdownStart))
	log.Printf("Final application metrics:\n  Start Time: %s\n  Total Uptime: %v",
		app.metrics.StartTime.Format(time.RFC3339), app.metrics.TotalUptime)
}

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
	cfg       *config.Config
	db        *database.Database
	webServer *http.Server
	router    *chi.Mux
	services  *services.Services
	handlers  *handlers.Handlers
	metrics   *models.ApplicationMetrics
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initialize application
	app, err := initializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.cleanup()

	// Setup graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		log.Printf("Received shutdown signal: %v", sig)
		log.Printf("Initiating graceful shutdown...")

		app.initiateShutdown()
	}()

	// Configure middleware
	app.router.Use(middleware.Logger)                    // Request logging
	app.router.Use(middleware.Recoverer)                 // Panic recovery
	app.router.Use(middleware.Compress(5))               // Response compression
	app.router.Use(middleware.Timeout(30 * time.Second)) // Request timeout

	// Configure routes
	app.setupRouter()

	// Configure server
	app.webServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", app.cfg.Host, app.cfg.Port),
		Handler:      app.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting web server on %s", app.webServer.Addr)

	// Start server (blocking call)
	if err = app.webServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Web server failed to start: %v", err)
	}
}

func (app *Application) setupRouter() {
	// Setup API routes
	app.handlers.EmailHandler.RegisterRoutes(app.router)
}

func initializeApplication() (*Application, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("configuration loading failed: %w", err)
	}

	time.Local = cfg.Timezone
	log.Printf("Application timezone: %s\n", cfg.Timezone)

	// Initialize database connection
	db, err := database.InitializeDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}

	// Initialize all services with dependency injection
	allServices := services.InitializeServices(db, cfg)

	// Initialize all handlers with service dependencies
	allHandlers := handlers.InitializeHandlers(allServices)

	app := &Application{
		cfg:      cfg,
		db:       db,
		router:   chi.NewRouter(),
		services: allServices,
		handlers: allHandlers,
		metrics:  &models.ApplicationMetrics{StartTime: time.Now()},
	}

	return app, nil
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
		app.metrics.StartTime.Format(time.RFC3339),
		app.metrics.TotalUptime)
}

// cleanup performs cleanup operations for application shutdown
func (app *Application) cleanup() {
	log.Printf("Performing application cleanup...")

	if app.db.Mongo != nil {
		log.Printf("Closing database connection...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.db.Mongo.Disconnect(ctx); err != nil {
			log.Printf("Database disconnection error: %v", err)
			return
		}
		log.Printf("Database connection closed successfully")
	}

	log.Printf("Application cleanup completed")
}

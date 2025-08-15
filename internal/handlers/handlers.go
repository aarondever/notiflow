package handlers

import (
	"github.com/aarondever/url-forg/internal/services"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	EmailHandler *EmailHandler
}

func InitializeHandlers(services *services.Services) *Handlers {
	// Initialize each handler with its service dependencies
	return &Handlers{
		EmailHandler: NewEmailHandler(services.URLService),
	}
}

func (handlers *Handlers) SetupRouters(router *chi.Mux) {
	// Setup API routes
	handlers.EmailHandler.RegisterRoutes(router)
}

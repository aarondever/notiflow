package handlers

import "github.com/aarondever/url-forg/internal/services"

type Handlers struct {
	EmailHandler *EmailHandler
}

func InitializeHandlers(services *services.Services) *Handlers {
	// Initialize each handler with its service dependencies
	return &Handlers{
		EmailHandler: NewEmailHandler(services.URLService),
	}
}

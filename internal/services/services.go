package services

import (
	"github.com/aarondever/url-forg/internal/config"
	"github.com/aarondever/url-forg/internal/database"
)

type Services struct {
	URLService *EmailService
}

func InitializeServices(db *database.Database, cfg *config.Config) *Services {
	// Initialize each service - add new services here
	return &Services{
		URLService: NewEmailService(db, cfg),
	}
}

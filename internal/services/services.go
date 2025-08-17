package services

import (
	"github.com/aarondever/notiflow/internal/config"
	"github.com/aarondever/notiflow/internal/database"
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

//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/aarondever/notiflow/internal/config"
	"github.com/aarondever/notiflow/internal/database"
	"github.com/aarondever/notiflow/internal/handlers"
	"github.com/aarondever/notiflow/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

type App struct {
	DB     *database.Database
	Router *gin.Engine
}

func NewApp(
	db *database.Database,
	emailHandler *handlers.EmailHandler,
	// Add all handlers as parameters
) *App {
	router := gin.Default()

	// Setup all routes
	emailHandler.SetupRouters(router)

	return &App{
		DB:     db,
		Router: router,
	}
}

// InitializeApp uses Wire to initialize all dependencies
func InitializeApp(cfg *config.Config) (*App, error) {
	wire.Build(
		database.NewDatabase,
		services.ProviderSet,
		handlers.ProviderSet,
		NewApp,
	)

	return nil, nil
}

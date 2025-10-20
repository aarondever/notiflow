//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/aarondever/notiflow/internal/config"
	"github.com/aarondever/notiflow/internal/database"
	"github.com/aarondever/notiflow/internal/grpc"
	"github.com/aarondever/notiflow/internal/handlers"
	"github.com/aarondever/notiflow/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	grpcServer "google.golang.org/grpc"
)

type App struct {
	DB         *database.Database
	Router     *gin.Engine
	GRPCServer *grpcServer.Server
}

func NewApp(
	db *database.Database,
	emailHandler *handlers.EmailHandler,
	emailGRPCHandler *grpc.EmailGRPCHandler,
	// Add all handlers as parameters
) *App {
	// Setup HTTP router
	router := gin.Default()
	emailHandler.SetupRouters(router)

	// Setup gRPC server
	grpcSrv := grpcServer.NewServer()
	grpc.RegisterEmailService(grpcSrv, emailGRPCHandler)

	return &App{
		DB:         db,
		Router:     router,
		GRPCServer: grpcSrv,
	}
}

// InitializeApp uses Wire to initialize all dependencies
func InitializeApp(cfg *config.Config) (*App, error) {
	wire.Build(
		database.NewDatabase,
		services.ProviderSet,
		handlers.ProviderSet,
		grpc.ProviderSet,
		NewApp,
	)

	return nil, nil
}

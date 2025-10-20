//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/aarondever/notiflow/internal/config"
	"github.com/aarondever/notiflow/internal/database"
	"github.com/aarondever/notiflow/internal/handlers"
	"github.com/aarondever/notiflow/internal/services"
	"github.com/aarondever/notiflow/proto/email"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

type App struct {
	DB         *database.Database
	Router     *gin.Engine
	GRPCServer *grpc.Server
}

func NewApp(
	db *database.Database,
	emailHandler *handlers.EmailHandler,
	emailGRPCHandler *handlers.EmailGRPCHandler,
	// Add all handlers as parameters
) *App {
	// Setup HTTP router
	router := gin.Default()
	emailHandler.RegisterRouter(router)

	// Setup gRPC server
	grpcSrv := grpc.NewServer()
	email.RegisterEmailServiceServer(grpcSrv, emailGRPCHandler)

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
		NewApp,
	)

	return nil, nil
}

package services

import (
	"context"
	"github.com/aarondever/url-forg/internal/config"
	"github.com/aarondever/url-forg/internal/database"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

const EmailCollectionName = "emails"

type EmailService struct {
	db              *database.Database
	cfg             *config.Config
	emailCollection *mongo.Collection
}

func NewEmailService(db *database.Database, cfg *config.Config) *EmailService {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.CreateCollection(ctx, EmailCollectionName); err != nil {
		return nil
	}

	collection := db.MongoDB.Collection(EmailCollectionName)

	return &EmailService{
		db:              db,
		cfg:             cfg,
		emailCollection: collection,
	}
}

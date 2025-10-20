package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aarondever/notiflow/internal/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Database struct {
	Mongo           *mongo.Client
	db              *mongo.Database
	emailCollection *mongo.Collection
}

func NewDatabase(config *config.Config) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	databaseURL := fmt.Sprintf("mongodb://%s:%s@%s:%d",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
	)

	// Connect to MongoDB
	client, err := mongo.Connect(options.Client().ApplyURI(databaseURL))
	if err != nil {
		slog.Error("Failed to connect to MongoDB", "error", err)
		return nil, err
	}

	// Test MongoDB connection
	if err = client.Ping(ctx, nil); err != nil {
		slog.Error("Failed to ping MongoDB", "error", err)
		client.Disconnect(ctx)
		return nil, err
	}

	slog.Info("Connected to MongoDB")

	database := &Database{
		Mongo: client,
		db:    client.Database(config.Database.Name),
	}

	// Initialize collections
	database.emailCollection = database.initEmailCollection(ctx)

	return database, nil
}

func (database *Database) createCollection(ctx context.Context, collectionName string, validator bson.M) {
	// If collection exists, skip creation
	collections, _ := database.db.ListCollectionNames(ctx, bson.M{"name": collectionName})
	if len(collections) > 0 {
		return
	}

	// Create collection with validation schema
	opts := options.CreateCollection().SetValidator(validator)
	if err := database.db.CreateCollection(ctx, collectionName, opts); err != nil {
		slog.Error("Failed to create collection", "collection", collectionName, "error", err)
		os.Exit(1)
	}

	slog.Info("Collection created successfully", "collection", collectionName)
}

func (database *Database) createIndexes(ctx context.Context, collection *mongo.Collection, indexes []mongo.IndexModel) {
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("Failed to create indexes", "collection", collection.Name(), "error", err)
		os.Exit(1)
	}

	slog.Info("Indexes created successfully", "collection", collection.Name())
}

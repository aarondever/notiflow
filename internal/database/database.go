package database

import (
	"context"
	"fmt"
	"github.com/aarondever/notiflow/internal/config"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log/slog"
	"os"
	"time"
)

type Database struct {
	Mongo           *mongo.Client
	db              *mongo.Database
	Redis           *redis.Client
	emailCollection *mongo.Collection
}

func InitializeDatabase(config *config.Config) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	databaseURL := fmt.Sprintf("mongodb://%s:%s@%s:%d",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port)

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

	// Connect to Redis
	//if err = database.connectToRedis(config); err != nil {
	//	client.Disconnect(ctx)
	//	return nil, err
	//}

	return database, nil
}

func (database *Database) connectToRedis(config *config.Config) error {
	var redisURL string
	// Format: redis://[:password@]host:port/db
	if config.Redis.Password != "" {
		redisURL = fmt.Sprintf("redis://:%s@%s:%d/%d",
			config.Redis.Password,
			config.Redis.Host,
			config.Redis.Port,
			config.Redis.DB)
	} else {
		redisURL = fmt.Sprintf("redis://%s:%d/%d",
			config.Redis.Host,
			config.Redis.Port,
			config.Redis.DB)
	}

	// Initialize Redis client
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		slog.Error("Error parsing Redis URL", "error", err)
		return err
	}

	database.Redis = redis.NewClient(opts)

	// Verify Redis connectivity
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPing()
	if err = database.Redis.Ping(ctxPing).Err(); err != nil {
		slog.Error("Redis ping failed", "error", err, "address", opts.Addr, "db", opts.DB)
		return err
	}

	slog.Info("Connected to Redis successfully", "address", opts.Addr, "db", opts.DB)

	return nil
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

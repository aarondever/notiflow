package database

import (
	"context"
	"fmt"
	"github.com/aarondever/url-forg/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log/slog"
	"time"
)

const emailCollectionName = "emails"

func (database *Database) GetEmailByID(ctx context.Context, id string) (*models.Email, error) {
	emailID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		slog.Error("Failed to parse email ID", "error", err)
		return nil, err
	}

	var email models.Email
	if err = database.emailCollection.FindOne(ctx, bson.M{"_id": emailID}).Decode(&email); err != nil {
		slog.Error("Failed to find email", "error", err)
		return nil, err
	}

	return &email, nil
}

func (database *Database) CreateEmail(ctx context.Context, email models.Email) (*models.Email, error) {
	email.CreatedAt = time.Now()
	email.Status = models.StatusPending

	result, err := database.emailCollection.InsertOne(ctx, email)
	if err != nil {
		slog.Error("Failed to insert email", "error", err)
		return nil, err
	}

	return database.GetEmailByID(ctx, result.InsertedID.(bson.ObjectID).Hex())
}

func (database *Database) UpdateEmail(ctx context.Context, email models.Email) (*models.Email, error) {
	if email.ID == bson.NilObjectID {
		return nil, fmt.Errorf("ID is required for updating an email")
	}

	_, err := database.emailCollection.UpdateOne(ctx, bson.M{"_id": email.ID}, email)
	if err != nil {
		slog.Error("Failed to update email", "error", err)
		return nil, err
	}

	return database.GetEmailByID(ctx, email.ID.Hex())
}

func (database *Database) initEmailCollection(ctx context.Context) *mongo.Collection {
	database.createCollection(ctx, emailCollectionName, bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"to", "subject", "body", "status", "created_at", "is_html"},
			"properties": bson.M{
				"to": bson.M{
					"bsonType": "array",
					"minItems": 1,
					"maxItems": 100,
					"items": bson.M{
						"bsonType": "string",
						"pattern":  "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
					},
					"description": "must be an array of valid email addresses and is required",
				},
				"cc": bson.M{
					"bsonType": "array",
					"maxItems": 50,
					"items": bson.M{
						"bsonType": "string",
						"pattern":  "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
					},
					"description": "must be an array of valid email addresses",
				},
				"bcc": bson.M{
					"bsonType": "array",
					"maxItems": 50,
					"items": bson.M{
						"bsonType": "string",
						"pattern":  "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
					},
					"description": "must be an array of valid email addresses",
				},
				"subject": bson.M{
					"bsonType":    "string",
					"minLength":   1,
					"maxLength":   255,
					"description": "must be a string between 1-255 characters and is required",
				},
				"body": bson.M{
					"bsonType":    "string",
					"minLength":   1,
					"maxLength":   1048576, // 1MB limit
					"description": "must be a string up to 1MB and is required",
				},
				"is_html": bson.M{
					"bsonType":    "bool",
					"description": "must be a boolean indicating if body is HTML",
				},
				"status": bson.M{
					"bsonType":    "string",
					"enum":        []string{"pending", "sent", "failed"},
					"description": "must be one of: pending, sent, failed",
				},
				"error_message": bson.M{
					"bsonType":    "string",
					"maxLength":   1000,
					"description": "must be a string up to 1000 characters",
				},
				"created_at": bson.M{
					"bsonType":    "date",
					"description": "must be a date and is required",
				},
				"sent_at": bson.M{
					"bsonType":    "date",
					"description": "must be a date",
				},
				"attachments": bson.M{
					"bsonType": "array",
					"maxItems": 10,
					"items": bson.M{
						"bsonType": "object",
						"required": []string{"filename", "content", "content_type"},
						"properties": bson.M{
							"filename": bson.M{
								"bsonType":    "string",
								"minLength":   1,
								"maxLength":   255,
								"description": "must be a string between 1-255 characters",
							},
							"content": bson.M{
								"bsonType":    "binData",
								"description": "must be binary data",
							},
							"content_type": bson.M{
								"bsonType":    "string",
								"minLength":   1,
								"maxLength":   100,
								"description": "must be a string between 1-100 characters",
							},
						},
					},
					"description": "must be an array of attachment objects (max 10)",
				},
			},
		},
	})

	collection := database.db.Collection(emailCollectionName)

	database.createIndexes(ctx, collection, []mongo.IndexModel{
		// Index on created_at for chronological queries
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_desc"),
		},
		// Index on status for filtering by email status
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("status_asc"),
		},
		// Index on to field for finding emails by recipient
		{
			Keys:    bson.D{{Key: "to", Value: 1}},
			Options: options.Index().SetName("to_asc"),
		},
		// Text index for full-text search on subject and body
		{
			Keys: bson.D{
				{Key: "subject", Value: "text"},
				{Key: "body", Value: "text"},
			},
			Options: options.Index().
				SetName("email_text_search").
				SetWeights(bson.D{
					{Key: "subject", Value: 10},
					{Key: "body", Value: 1},
				}),
		},
		// TTL Index for automatic cleanup of old emails (optional - 90 days)
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().
				SetName("email_ttl").
				SetExpireAfterSeconds(90 * 24 * 60 * 60), // 90 days
		},
	})

	return collection
}

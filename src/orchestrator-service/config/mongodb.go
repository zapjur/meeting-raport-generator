package config

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	retries    = 5
	retryDelay = 5 * time.Second
)

func ConnectToMongoDB(uri string) (*mongo.Client, error) {
	var client *mongo.Client
	var err error

	for i := 0; i < retries; i++ {
		client, err = mongo.NewClient(options.Client().ApplyURI(uri))
		if err != nil {
			log.Printf("MongoDB client creation error: %v", err)
			time.Sleep(retryDelay)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.Connect(ctx)
		if err == nil {
			log.Println("Successfully connected to MongoDB.")
			return client, nil
		}

		log.Printf("MongoDB connection failed (%d/%d): %v", i+1, retries, err)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("failed to connect to MongoDB after %d retries: %w", retries, err)
}

func DeleteMeetingData(ctx context.Context, client *mongo.Client, databaseName, meetingID string) error {
	collections := []string{"summaries", "ocr_results", "transcriptions", "embeddings"}

	for _, collectionName := range collections {
		collection := client.Database(databaseName).Collection(collectionName)

		filter := bson.M{"meeting_id": meetingID}

		result, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			log.Printf("Error deleting documents from collection %s: %v", collectionName, err)
			return err
		}

		log.Printf("Deleted %d documents from collection %s with meeting_id=%s", result.DeletedCount, collectionName, meetingID)
	}

	return nil
}

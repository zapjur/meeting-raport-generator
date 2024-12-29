package main

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Config struct {
	MongoClient   *mongo.Client
	RabbitConn    *amqp.Connection
	RabbitChannel *amqp.Channel
}

// Retry logic for MongoDB connection
func connectToMongoDB(uri string, retries int, delay time.Duration) (*mongo.Client, error) {
	var client *mongo.Client
	var err error

	for i := 0; i < retries; i++ {
		client, err = mongo.NewClient(options.Client().ApplyURI(uri))
		if err != nil {
			log.Printf("MongoDB client creation error: %v", err)
			time.Sleep(delay)
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
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("failed to connect to MongoDB after %d retries: %w", retries, err)
}

// Retry logic for RabbitMQ connection
func connectToRabbitMQ(uri string, retries int, delay time.Duration) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < retries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ.")
			return conn, nil
		}

		log.Printf("RabbitMQ connection failed (%d/%d): %v", i+1, retries, err)
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("failed to connect to RabbitMQ after %d retries: %w", retries, err)
}

func main() {
	const retries = 5
	const retryDelay = 5 * time.Second

	// Connect to MongoDB
	mongoURI := "mongodb://admin:password@mongodb:27017"
	mongoClient, err := connectToMongoDB(mongoURI, retries, retryDelay)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Connect to RabbitMQ
	//rabbitMQURI := "amqp://guest:guest@rabbitmq:5672/"
	//rabbitConn, err := connectToRabbitMQ(rabbitMQURI, retries, retryDelay)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer rabbitConn.Close()
	//
	//rabbitChannel, err := rabbitConn.Channel()
	//if err != nil {
	//	log.Fatal("RabbitMQ channel error:", err)
	//}
	//defer rabbitChannel.Close()

	app := &Config{
		MongoClient: mongoClient,
		//RabbitConn:    rabbitConn,
		//RabbitChannel: rabbitChannel,
	}

	app.generateSummary()
}

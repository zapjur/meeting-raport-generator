package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogMessage struct {
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
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
	// Retry settings
	const retries = 5
	const retryDelay = 5 * time.Second

	// Connect to MongoDB
	mongoURI := "mongodb://admin:password@mongodb:27017"
	client, err := connectToMongoDB(mongoURI, retries, retryDelay)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("logger").Collection("logs")

	// Connect to RabbitMQ
	rabbitMQURI := "amqp://guest:guest@rabbitmq:5672/"
	conn, err := connectToRabbitMQ(rabbitMQURI, retries, retryDelay)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("RabbitMQ channel error:", err)
	}
	defer channel.Close()

	msgs, err := channel.Consume(
		"logs_queue", // Queue name
		"",           // Consumer name
		true,         // Auto-ack
		false,        // Exclusive
		false,        // No-local
		false,        // No-wait
		nil,          // Args
	)
	if err != nil {
		log.Fatal("RabbitMQ consume error:", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var logMessage LogMessage
			err = json.Unmarshal(d.Body, &logMessage)
			if err != nil {
				log.Println("JSON Unmarshal error:", err)
				continue
			}

			// Insert log into MongoDB
			_, err = collection.InsertOne(context.Background(), logMessage)
			if err != nil {
				log.Println("MongoDB insert error:", err)
			} else {
				fmt.Println("Log saved:", logMessage)
			}
		}
	}()

	fmt.Println("Logger Service is listening to RabbitMQ...")
	<-forever
}

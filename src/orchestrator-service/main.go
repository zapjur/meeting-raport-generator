package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	retries    = 5
	retryDelay = 5 * time.Second
	webPort    = "8080"
)

type Config struct {
	MongoClient   *mongo.Client
	RabbitConn    *amqp.Connection
	RabbitChannel *amqp.Channel
}

func connectToMongoDB(uri string) (*mongo.Client, error) {
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

func connectToRabbitMQ(uri string) (*amqp.Connection, *amqp.Channel, error) {
	var conn *amqp.Connection
	var channel *amqp.Channel
	var err error

	for i := 0; i < retries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			channel, err = conn.Channel()
			if err == nil {
				log.Println("Successfully connected to RabbitMQ.")
				return conn, channel, nil
			}
		}

		log.Printf("RabbitMQ connection failed (%d/%d): %v", i+1, retries, err)
		time.Sleep(retryDelay)
	}

	return nil, nil, fmt.Errorf("failed to connect to RabbitMQ after %d retries: %w", retries, err)
}

func main() {
	// MongoDB connection
	mongoURI := "mongodb://admin:password@mongodb:27017"
	mongoClient, err := connectToMongoDB(mongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())

	// RabbitMQ connection
	rabbitMQURI := "amqp://guest:guest@rabbitmq:5672/"
	rabbitConn, rabbitChannel, err := connectToRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	// Declare necessary queues
	queues := []string{"logs_queue", "summary_queue"}
	for _, queue := range queues {
		_, err = rabbitChannel.QueueDeclare(
			queue, // queue name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			log.Fatalf("Failed to declare RabbitMQ queue '%s': %v", queue, err)
		}
		log.Printf("Queue '%s' declared successfully.", queue)
	}

	app := &Config{
		MongoClient:   mongoClient,
		RabbitConn:    rabbitConn,
		RabbitChannel: rabbitChannel,
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	if err = app.sendSummaryTask("867297"); err != nil {
		log.Printf("Error sending summary task: %v", err)
	}

	log.Printf("Orchestrator is ready and listening on port %s.", webPort)
	if err = srv.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

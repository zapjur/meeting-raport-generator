package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"orchestrator-service/config"
	"orchestrator-service/handlers"
	"orchestrator-service/rabbitmq"
	"orchestrator-service/redis"
	"orchestrator-service/routes"
)

const webPort = "8080"

func main() {
	mongoURI := "mongodb://admin:password@mongodb:27017"
	mongoClient, err := config.ConnectToMongoDB(mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	rabbitMQURI := "amqp://guest:guest@rabbitmq:5672/"
	rabbitConn, rabbitChannel, err := config.ConnectToRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	queues := []string{"logs_queue", "summary_queue", "transcription_queue", "ocr_queue", "report_queue", "orchestrator_ack_queue", "email_queue"}
	err = rabbitmq.DeclareQueues(rabbitChannel, queues)
	if err != nil {
		log.Fatalf("Failed to declare RabbitMQ queues: %v", err)
	}

	redisClient, err := config.ConnectToRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	redisManager := &redis.RedisManager{Client: redisClient}

	taskHandler := &handlers.TaskHandler{
		RabbitChannel: rabbitChannel,
		RedisManager:  redisManager,
	}

	rabbitConsumer := &rabbitmq.RabbitMQConsumer{Channel: rabbitChannel}
	err = rabbitConsumer.Consume("orchestrator_ack_queue", handlers.HandleAckMessage(redisManager, taskHandler))
	if err != nil {
		log.Fatalf("Failed to start RabbitMQ consumer: %v", err)
	}

	r := routes.Routes(&handlers.Config{
		MongoClient:   mongoClient,
		RabbitChannel: rabbitChannel,
		RedisManager:  &redis.RedisManager{Client: redisClient},
		TaskHandler:   taskHandler,
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: r,
	}

	log.Printf("Orchestrator is ready and listening on port %s.", webPort)
	if err = srv.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

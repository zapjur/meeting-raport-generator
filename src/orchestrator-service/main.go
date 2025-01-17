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

	queues := []string{"logs_queue", "summary_queue", "transcription_queue", "ocr_queue", "report_queue", "orchestrator_ack_queue"}
	err = rabbitmq.DeclareQueues(rabbitChannel, queues)
	if err != nil {
		log.Fatalf("Failed to declare RabbitMQ queues: %v", err)
	}

	redisClient, err := config.ConnectToRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	taskHandler := &handlers.TaskHandler{
		RabbitChannel: rabbitChannel,
		RedisManager:  &redis.RedisManager{Client: redisClient},
	}

	//err = taskHandler.SendSummaryTask("867297")
	//if err != nil {
	//	log.Printf("Error sending summary task: %v", err)
	//}

	//err = taskHandler.SendTranscriptionTask("867297")
	//if err != nil {
	//	log.Printf("Error sending transcription task: %v", err)
	//}

	//err = taskHandler.SendOcrTask("867297")
	//if err != nil {
	//	log.Printf("Error sending OCR task: %v", err)
	//}

	err = taskHandler.SendReportTask("867297")
	if err != nil {
		log.Printf("Error sending report task: %v", err)
	}

	r := routes.Routes(&handlers.Config{
		MongoClient:   mongoClient,
		RabbitChannel: rabbitChannel,
		RedisManager:  redisClient,
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

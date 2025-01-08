package main

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

type LogMessage struct {
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
}

func (app *Config) processSummaryTasks() {
	msgs, err := app.RabbitChannel.Consume(
		"summary_queue", // queue name
		"",              // consumer tag
		true,            // auto-acknowledge
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Println("Listening for messages on summary_queue...")

	for msg := range msgs {
		log.Printf("Received task: %s", msg.Body)

		var task map[string]string
		if err = json.Unmarshal(msg.Body, &task); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		meetingId, ok := task["meeting_id"]
		if !ok {
			log.Println("Invalid task message: missing 'meeting_id'")
			continue
		}

		log.Printf("Processing task with meeting_id: %s", meetingId)

		logMessage := LogMessage{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Service:   "summary",
			Level:     "INFO",
			Message:   fmt.Sprintf("Processing summary task with meeting_id: %s", meetingId),
			Details: map[string]interface{}{
				"meeting_id": meetingId,
				"queue":      "summary_queue",
			},
		}
		if err = app.publishLog(logMessage); err != nil {
			log.Printf("Error publishing log to RabbitMQ: %v", err)
		}

		app.generateSummary(meetingId)
	}
}

func (app *Config) publishLog(logMessage LogMessage) error {
	body, err := json.Marshal(logMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal log message: %w", err)
	}

	err = app.RabbitChannel.Publish(
		"",           // exchange
		"logs_queue", // routing key (kolejka docelowa)
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish log: %w", err)
	}

	log.Println("Log sent to RabbitMQ:", string(body))
	return nil
}

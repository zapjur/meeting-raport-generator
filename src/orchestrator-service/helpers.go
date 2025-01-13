package main

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type LogMessage struct {
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
}

func (app *Config) sendSummaryTask(meetingId string) error {
	taskMessage := fmt.Sprintf(`{"meeting_id": "%s"}`, meetingId)
	err := app.RabbitChannel.Publish(
		"",              // exchange
		"summary_queue", // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(taskMessage),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}
	log.Printf("Task sent to summary_queue: %s", taskMessage)
	return nil
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

func (app *Config) sendTranscriptionTask(meetingId string) error {
	filePath := "/shared/test_audio.wav"
	taskMessage := fmt.Sprintf(`{"meeting_id": "%s", "file_path": "%s"}`, meetingId, filePath)
	err := app.RabbitChannel.Publish(
		"",                    // exchange
		"transcription_queue", // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(taskMessage),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}
	log.Printf("Task sent to transcription_queue: %s", taskMessage)
	return nil
}

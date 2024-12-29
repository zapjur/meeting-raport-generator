package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

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

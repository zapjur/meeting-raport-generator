package main

import (
	"encoding/json"
	"log"
)

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
		app.generateSummary(meetingId)
	}
}

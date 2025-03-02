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

type AckMessage struct {
	MeetingId string `json:"meeting_id"`
	TaskId    string `json:"task_id"`
	TaskType  string `json:"task_type"`
	Status    string `json:"status"`
}

type TaskMessage struct {
	MeetingId string `json:"meeting_id"`
	TaskId    string `json:"task_id"`
}

func (app *Config) sendAckMessage(meetingId, taskId, status string) error {
	ackMessage := AckMessage{
		MeetingId: meetingId,
		TaskId:    taskId,
		TaskType:  "report",
		Status:    status,
	}

	body, err := json.Marshal(ackMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal ack message: %w", err)
	}

	err = app.RabbitChannel.Publish(
		"",                       // exchange
		"orchestrator_ack_queue", // routing key (kolejka docelowa)
		false,                    // mandatory
		false,                    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish ack message: %w", err)
	}

	log.Printf("Ack message sent to orchestrator_ack_queue: %s", string(body))
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

func (app *Config) processReportTasks() {
	for {
		_, err := app.RabbitChannel.QueueDeclare(
			"report_queue", // queue name
			true,           // durable
			false,          // delete when unused
			false,          // exclusive
			false,          // no-wait
			nil,            // arguments
		)
		if err != nil {
			log.Printf("Failed to declare queue: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		msgs, err := app.RabbitChannel.Consume(
			"report_queue", // queue name
			"",             // consumer tag
			false,          // auto-acknowledge
			false,          // exclusive
			false,          // no-local
			false,          // no-wait
			nil,            // arguments
		)
		if err != nil {
			log.Printf("Failed to register consumer: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Listening for messages on report_queue...")

		for msg := range msgs {
			log.Printf("Received task: %s", msg.Body)

			var task TaskMessage
			if err = json.Unmarshal(msg.Body, &task); err != nil {
				log.Printf("Error parsing message: %v", err)
				_ = msg.Nack(false, false)
				continue
			}

			if task.MeetingId == "" {
				log.Println("Invalid task message: missing 'meeting_id'")
				_ = msg.Nack(false, false)
			}

			log.Printf("Processing task with meeting_id: %s", task.MeetingId)

			transcriptions, summary, ocrResults, err := app.fetchMeetingData(task.MeetingId)
			if err != nil {
				log.Printf("Error fetching data for meeting_id %s: %v", task.MeetingId, err)
				_ = app.sendAckMessage(task.MeetingId, task.TaskId, "failed")
				_ = msg.Nack(false, false)
				continue
			}

			screenshots, err := fetchScreenshots(task.MeetingId)
			if err != nil {
				log.Printf("Error fetching screenshots for meeting_id %s: %v", task.MeetingId, err)
				_ = app.sendAckMessage(task.MeetingId, task.TaskId, "failed")
				_ = msg.Nack(false, false)
				continue
			}

			err = generatePDF(task.MeetingId, transcriptions, summary, ocrResults, screenshots)
			if err != nil {
				log.Printf("Error generating PDF for meeting_id %s: %v", task.MeetingId, err)
				_ = app.sendAckMessage(task.MeetingId, task.TaskId, "failed")
				_ = msg.Nack(false, false)
				continue
			}

			log.Printf("Successfully generated PDF for meeting_id: %s", task.MeetingId)

			err = app.sendAckMessage(task.MeetingId, task.TaskId, "completed")
			if err != nil {
				log.Printf("Error sending acknowledgment message for meeting_id %s: %v", task.MeetingId, err)
			}

			_ = msg.Ack(false)
		}

		log.Println("Message loop ended. Reconnecting...")
	}
}

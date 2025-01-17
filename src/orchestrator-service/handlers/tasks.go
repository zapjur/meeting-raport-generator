package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"orchestrator-service/redis"
	"time"

	"github.com/streadway/amqp"
)

type TaskHandler struct {
	RabbitChannel *amqp.Channel
	RedisManager  *redis.RedisManager
}

type LogMessage struct {
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
}

func (h *TaskHandler) sendReportTask(meetingId string) error {
	taskID := fmt.Sprintf("%s-report-%d", meetingId, time.Now().UnixNano())
	taskMessage := fmt.Sprintf(`{"meeting_id": "%s", "task_id": "%s"}`, meetingId, taskID)

	ctx := context.Background()
	err := h.RedisManager.AddTask(ctx, meetingId, taskID)
	if err != nil {
		return fmt.Errorf("failed to add task to Redis: %w", err)
	}

	err = h.RabbitChannel.Publish(
		"",
		"report_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          []byte(taskMessage),
			CorrelationId: taskID,
		},
	)
	if err != nil {
		h.RedisManager.UpdateTaskStatus(ctx, meetingId, taskID, "failed")
		return fmt.Errorf("failed to publish task to RabbitMQ: %w", err)
	}

	log.Printf("Task sent to report_queue: %s", taskMessage)
	return nil
}

func (h *TaskHandler) SendSummaryTask(meetingId string) error {
	taskID := fmt.Sprintf("%s-summary-%d", meetingId, time.Now().UnixNano())
	taskMessage := fmt.Sprintf(`{"meeting_id": "%s", "task_id": "%s"}`, meetingId, taskID)

	ctx := context.Background()
	err := h.RedisManager.AddTask(ctx, meetingId, taskID)
	if err != nil {
		return fmt.Errorf("failed to add task to Redis: %w", err)
	}

	err = h.RabbitChannel.Publish(
		"",
		"summary_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          []byte(taskMessage),
			CorrelationId: taskID,
		},
	)
	if err != nil {
		h.RedisManager.UpdateTaskStatus(ctx, meetingId, taskID, "failed")
		return fmt.Errorf("failed to publish task to RabbitMQ: %w", err)
	}

	log.Printf("Task sent to summary_queue: %s", taskMessage)
	return nil
}

func (h *TaskHandler) SendTranscriptionTask(meetingId string) error {
	taskID := fmt.Sprintf("%s-transcription-%d", meetingId, time.Now().UnixNano())
	filePath := "/shared-transcription/test_audio.wav"
	taskMessage := fmt.Sprintf(`{"meeting_id": "%s", "file_path": "%s", "task_id": "%s"}`, meetingId, filePath, taskID)

	ctx := context.Background()
	err := h.RedisManager.AddTask(ctx, meetingId, taskID)
	if err != nil {
		return fmt.Errorf("failed to add task to Redis: %w", err)
	}

	err = h.RabbitChannel.Publish(
		"",
		"transcription_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          []byte(taskMessage),
			CorrelationId: taskID,
		},
	)
	if err != nil {
		h.RedisManager.UpdateTaskStatus(ctx, meetingId, taskID, "failed")
		return fmt.Errorf("failed to publish task to RabbitMQ: %w", err)
	}

	log.Printf("Task sent to transcription_queue: %s", taskMessage)
	return nil
}

func (h *TaskHandler) SendOcrTask(meetingId string) error {
	taskID := fmt.Sprintf("%s-ocr-%d", meetingId, time.Now().UnixNano())
	filePath := "/shared-ocr/test_ocr.png"
	taskMessage := fmt.Sprintf(`{"meeting_id": "%s", "file_path": "%s", "task_id": "%s"}`, meetingId, filePath, taskID)

	ctx := context.Background()
	err := h.RedisManager.AddTask(ctx, meetingId, taskID)
	if err != nil {
		return fmt.Errorf("failed to add task to Redis: %w", err)
	}

	err = h.RabbitChannel.Publish(
		"",
		"ocr_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          []byte(taskMessage),
			CorrelationId: taskID,
		},
	)
	if err != nil {
		h.RedisManager.UpdateTaskStatus(ctx, meetingId, taskID, "failed")
		return fmt.Errorf("failed to publish task to RabbitMQ: %w", err)
	}

	log.Printf("Task sent to ocr_queue: %s", taskMessage)
	return nil
}

func (h *TaskHandler) publishLog(logMessage LogMessage) error {
	body, err := json.Marshal(logMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal log message: %w", err)
	}

	err = h.RabbitChannel.Publish(
		"",           // exchange
		"logs_queue", // routing key
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

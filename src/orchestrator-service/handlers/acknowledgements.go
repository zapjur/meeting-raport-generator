package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
	"orchestrator-service/redis"
)

type AckMessage struct {
	MeetingId string `json:"meeting_id"`
	TaskId    string `json:"task_id"`
	TaskType  string `json:"task_type"`
	Status    string `json:"status"`
}

func HandleAckMessage(rm *redis.RedisManager) func(amqp.Delivery) {
	return func(msg amqp.Delivery) {
		var ack AckMessage
		if err := json.Unmarshal(msg.Body, &ack); err != nil {
			log.Printf("Failed to unmarshal ACK message: %v", err)
			return
		}

		log.Printf("Processing ACK: %+v", ack)

		ctx := context.Background()
		err := rm.UpdateTaskStatus(ctx, ack.MeetingId, ack.TaskId, ack.Status)
		if err != nil {
			log.Printf("Failed to update Redis for ACK: %v", err)
			return
		}
	}
}

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

func HandleAckMessage(rm *redis.RedisManager, taskHandler *TaskHandler) func(amqp.Delivery) {
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

		meetingStatus, err := rm.GetMeetingStatus(ctx, ack.MeetingId)
		if err != nil || meetingStatus != "ended" {
			log.Printf("Meeting %s is not ended yet", ack.MeetingId)
			return
		}

		if ack.TaskType == "transcription" && ack.Status != "pending" {
			allTranscriptionCompleted, err := rm.AllTasksOfTypeCompleted(ctx, ack.MeetingId, ack.TaskType)
			if err != nil {
				log.Printf("Failed to check if all tasks are completed for meeting_id: %s %v", ack.MeetingId, err)
				return
			}

			if allTranscriptionCompleted {
				log.Printf("All tasks completed for meeting_id: %s", ack.MeetingId)
				err = taskHandler.SendSummaryTask(ack.MeetingId)
				if err != nil {
					log.Printf("Error sending summary task for meeting_id: %s %v", ack.MeetingId, err)
				}
			}
		}
	}
}

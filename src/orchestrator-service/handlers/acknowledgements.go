package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
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

		if ack.TaskType == "ocr" || ack.TaskType == "summary" {
			err = checkAndSendReport(ctx, rm, taskHandler, ack.MeetingId)
			if err != nil {
				log.Printf("Error sending report task for meeting_id: %s %v", ack.MeetingId, err)
			}
		}

		if ack.TaskType == "report" && ack.Status != "pending" {
			log.Printf("All tasks completed for meeting_id: %s", ack.MeetingId)
			email, err := rm.GetMeetingEmail(ctx, ack.MeetingId)
			if err != nil {
				log.Printf("Failed to get email for meeting_id: %s %v", ack.MeetingId, err)
				return
			}
			filePath := fmt.Sprintf("/shared-report/%s/meeting_report_%s.pdf", ack.MeetingId, ack.MeetingId)
			err = taskHandler.SendEmailTask(ack.MeetingId, filePath, email)
			if err != nil {
				log.Printf("Error sending email task for meeting_id: %s %v", ack.MeetingId, err)
			}

		}

	}
}

func checkAndSendReport(ctx context.Context, rm *redis.RedisManager, taskHandler *TaskHandler, meetingID string) error {
	allOCRCompleted, err := rm.AllTasksOfTypeCompleted(ctx, meetingID, "ocr")
	if err != nil {
		return fmt.Errorf("failed to check OCR tasks for meeting_id: %s %v", meetingID, err)
	}

	allSummaryCompleted, err := rm.AllTasksOfTypeCompleted(ctx, meetingID, "summary")
	if err != nil {
		return fmt.Errorf("failed to check Summary tasks for meeting_id: %s %v", meetingID, err)
	}

	allTranscriptionCompleted, err := rm.AllTasksOfTypeCompleted(ctx, meetingID, "transcription")
	if err != nil {
		return fmt.Errorf("failed to check Transcription tasks for meeting_id: %s %v", meetingID, err)
	}

	if allOCRCompleted && allSummaryCompleted && allTranscriptionCompleted {
		log.Printf("All OCR and Summary tasks completed for meeting_id: %s. Sending report task...", meetingID)
		return taskHandler.SendReportTask(meetingID)
	}

	return nil
}

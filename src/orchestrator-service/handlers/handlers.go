package handlers

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"orchestrator-service/redis"
)

type MeetingIdResponse struct {
	MeetingId string `json:"meeting_id"`
}

type Config struct {
	MongoClient   *mongo.Client
	RabbitChannel *amqp.Channel
	RedisManager  *redis.RedisManager
	TaskHandler   *TaskHandler
}

func (app *Config) GenerateMeetingId(w http.ResponseWriter, r *http.Request) {
	//src := rand.NewSource(time.Now().UnixNano())
	//rng := rand.New(src)
	//
	//const idLength = 10
	//const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	//
	//meetingId := make([]byte, idLength)
	//for i := range meetingId {
	//	meetingId[i] = charset[rng.Intn(len(charset))]
	//}
	//test purpose
	meetingId := "867297"

	response := MeetingIdResponse{
		MeetingId: string(meetingId),
	}

	ctx := context.Background()
	if err := app.RedisManager.SetMeetingStatus(ctx, response.MeetingId, "started"); err != nil {
		http.Error(w, "Failed to set meeting status", http.StatusInternalServerError)
		log.Printf("Failed to set meeting status: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (app *Config) EndMeeting(w http.ResponseWriter, r *http.Request) {
	meetingId := r.URL.Query().Get("meeting_id")
	if meetingId == "" {
		http.Error(w, "Missing meeting_id query parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := app.RedisManager.SetMeetingStatus(ctx, meetingId, "ended"); err != nil {
		http.Error(w, "Failed to set meeting status", http.StatusInternalServerError)
		log.Printf("Failed to set meeting status: %v", err)
	}
	app.CheckDependenciesAndTriggerTasks(meetingId)

	w.WriteHeader(http.StatusOK)
}

func (app *Config) CaptureScreenshots(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) CaptureAudio(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) StartTranscription(w http.ResponseWriter, r *http.Request) {
	err := app.TaskHandler.SendTranscriptionTask("867297")
	if err != nil {
		log.Printf("Error sending transcription task: %v", err)
	}
}

func (app *Config) StartOcr(w http.ResponseWriter, r *http.Request) {
	err := app.TaskHandler.SendOcrTask("867297")
	if err != nil {
		log.Printf("Error sending OCR task: %v", err)
	}
}

func (app *Config) CheckDependenciesAndTriggerTasks(meetingId string) {
	ctx := context.Background()

	allTranscriptionCompleted, err := app.RedisManager.AllTasksOfTypeCompleted(ctx, meetingId, "transcription")
	if err != nil {
		log.Printf("Failed to check transcription tasks for meeting_id=%s: %v", meetingId, err)
		return
	}

	if allTranscriptionCompleted {
		log.Printf("All transcription tasks completed for meeting_id=%s. Sending summary task...", meetingId)
		err = app.TaskHandler.SendSummaryTask(meetingId)
		if err != nil {
			log.Printf("Error sending summary task for meeting_id=%s: %v", meetingId, err)
		}
	}
}

package handlers

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"net/http"
	"orchestrator-service/redis"
	"os"
	"path/filepath"
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
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		log.Printf("Error parsing form: %v", err)
		return
	}

	file, header, err := r.FormFile("screenshot")
	if err != nil {
		http.Error(w, "Missing screenshot file", http.StatusBadRequest)
		log.Printf("Error retrieving file: %v", err)
		return
	}
	defer file.Close()

	meetingId := r.FormValue("meeting_id")
	if meetingId == "" {
		http.Error(w, "Missing meeting_id", http.StatusBadRequest)
		return
	}

	basePath := "/shared-ocr"
	meetingDir := filepath.Join(basePath, meetingId)

	err = os.MkdirAll(meetingDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Unable to create directory", http.StatusInternalServerError)
		log.Printf("Error creating directory: %v", err)
		return
	}

	filePath := filepath.Join(meetingDir, header.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		log.Printf("Error saving file: %v", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		log.Printf("Error writing file: %v", err)
		return
	}

	log.Printf("Screenshot saved successfully in: %s", filePath)

	err = app.TaskHandler.SendOcrTask(meetingId, filePath)
	if err != nil {
		log.Printf("Error sending OCR task: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Screenshot uploaded successfully"))
}

func (app *Config) CaptureAudio(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) StartTranscription(w http.ResponseWriter, r *http.Request) {
	err := app.TaskHandler.SendTranscriptionTask("867297")
	if err != nil {
		log.Printf("Error sending transcription task: %v", err)
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

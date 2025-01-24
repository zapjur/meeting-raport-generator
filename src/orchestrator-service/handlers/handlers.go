package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"math/rand"
	"net/http"
	"orchestrator-service/redis"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func (app *Config) GenerateMeetingId(w http.ResponseWriter, r *http.Request) {

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Missing email in query params", http.StatusBadRequest)
		return
	}

	if !isValidEmail(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)

	const idLength = 10
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	meetingId := make([]byte, idLength)
	for i := range meetingId {
		meetingId[i] = charset[rng.Intn(len(charset))]
	}

	response := MeetingIdResponse{
		MeetingId: string(meetingId),
	}

	ctx := context.Background()
	if err := app.RedisManager.SetMeetingStatus(ctx, response.MeetingId, "started"); err != nil {
		http.Error(w, "Failed to set meeting status", http.StatusInternalServerError)
		log.Printf("Failed to set meeting status: %v", err)
	}

	if err := app.RedisManager.SetMeetingEmail(ctx, response.MeetingId, email); err != nil {
		http.Error(w, "Failed to set meeting email", http.StatusInternalServerError)
		log.Printf("Failed to set meeting email: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (app *Config) EndMeeting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var requestData struct {
		MeetingId string `json:"meeting_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		log.Printf("Error decoding JSON: %v", err)
		return
	}

	meetingId := requestData.MeetingId
	if meetingId == "" {
		http.Error(w, "Missing meeting_id in request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err = app.RedisManager.SetMeetingStatus(ctx, meetingId, "ended"); err != nil {
		http.Error(w, "Failed to set meeting status", http.StatusInternalServerError)
		log.Printf("Failed to set meeting status: %v", err)
		return
	}

	app.CheckDependenciesAndTriggerTasks(meetingId)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Meeting ended successfully"}`))
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
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		log.Printf("Error parsing form: %v", err)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "Missing audio file", http.StatusBadRequest)
		log.Printf("Error retrieving file: %v", err)
		return
	}
	defer file.Close()

	meetingId := r.FormValue("meeting_id")
	if meetingId == "" {
		http.Error(w, "Missing meeting_id", http.StatusBadRequest)
		return
	}

	basePath := "/shared-transcription"
	meetingDir := filepath.Join(basePath, meetingId)

	err = os.MkdirAll(meetingDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Unable to create directory", http.StatusInternalServerError)
		log.Printf("Error creating directory: %v", err)
		return
	}

	webmFilePath := filepath.Join(meetingDir, header.Filename)
	webmFile, err := os.Create(webmFilePath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		log.Printf("Error saving file: %v", err)
		return
	}
	defer webmFile.Close()

	_, err = io.Copy(webmFile, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		log.Printf("Error writing file: %v", err)
		return
	}

	log.Printf("WebM audio saved successfully in: %s", webmFilePath)

	wavFilePath := filepath.Join(meetingDir, strings.TrimSuffix(header.Filename, ".webm")+".wav")

	err = convertWebmToWav(webmFilePath, wavFilePath)
	if err != nil {
		http.Error(w, "Error converting audio to WAV", http.StatusInternalServerError)
		log.Printf("Error converting audio: %v", err)
		return
	}

	log.Printf("Audio converted to WAV successfully: %s", wavFilePath)

	err = os.Remove(webmFilePath)
	if err != nil {
		log.Printf("Error deleting original WebM file: %v", err)
	}

	log.Printf("Audio saved successfully in: %s", wavFilePath)

	err = app.TaskHandler.SendTranscriptionTask(meetingId, wavFilePath)
	if err != nil {
		log.Printf("Error sending transcription task: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Audio uploaded successfully"))
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

func convertWebmToWav(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-vn", "-acodec", "pcm_s16le", "-ar", "44100", "-ac", "2", outputPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}

	return nil
}

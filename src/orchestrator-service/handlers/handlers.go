package handlers

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"math/rand"
	"net/http"
	"time"
)

type MeetingIdResponse struct {
	MeetingId string `json:"meeting_id"`
}

type Config struct {
	MongoClient   *mongo.Client
	RabbitChannel *amqp.Channel
	RedisManager  *redis.Client
	TaskHandler   *TaskHandler
}

func (app *Config) GenerateMeetingId(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (app *Config) CaptureScreenshots(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) CaptureAudio(w http.ResponseWriter, r *http.Request) {

}

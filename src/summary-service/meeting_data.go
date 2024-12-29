package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

// Transcription struct represents a single document in the collection
type Transcription struct {
	ID             string `bson:"_id"`
	SpeakerID      string `bson:"speaker_id"`
	Transcription  string `bson:"transcription"`
	TimestampStart string `bson:"timestamp_start"`
	TimestampEnd   string `bson:"timestamp_end"`
	MeetingID      string `bson:"meeting_id"`
}

func (app *Config) fetchTranscriptions(meetingId string) (string, error) {
	database := app.MongoClient.Database("database")
	collection := database.Collection("transcriptions")

	filter := bson.M{"meeting_id": meetingId}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp_start", Value: 1}})
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return "", err
	}
	defer cursor.Close(context.TODO())

	var transcriptions []Transcription
	if err = cursor.All(context.TODO(), &transcriptions); err != nil {
		return "", err
	}

	var combinedTranscription strings.Builder
	for _, t := range transcriptions {
		if t.Transcription != "" {
			combinedTranscription.WriteString(t.Transcription)
			combinedTranscription.WriteString(" ")
		}
	}

	return combinedTranscription.String(), nil

}

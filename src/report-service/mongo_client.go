package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Transcription struct {
	SpeakerID      string `bson:"speaker_id"`
	Transcription  string `bson:"transcription"`
	TimestampStart string `bson:"timestamp_start"`
	TimestampEnd   string `bson:"timestamp_end"`
	MeetingID      string `bson:"meeting_id"`
}

type Summary struct {
	MeetingID   string `bson:"meeting_id"`
	SummaryText string `bson:"summary_text"`
}

type OCRResult struct {
	TextResult string `bson:"text_result"`
	MeetingID  string `bson:"meeting_id"`
}

func (app *Config) fetchMeetingData(meetingID string) ([]Transcription, Summary, []OCRResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database := app.MongoClient.Database("database")

	log.Printf("Fetching transcriptions for meeting_id: %s", meetingID)
	transcriptionsColl := database.Collection("transcriptions")
	var transcriptions []Transcription
	cursor, err := transcriptionsColl.Find(ctx, bson.M{"meeting_id": meetingID}, options.Find().SetSort(bson.M{"timestamp_start": 1}))
	if err != nil {
		return nil, Summary{}, nil, fmt.Errorf("error fetching transcriptions: %w", err)
	}
	if err = cursor.All(ctx, &transcriptions); err != nil {
		return nil, Summary{}, nil, fmt.Errorf("error decoding transcriptions: %w", err)
	}
	log.Printf("Fetched %d transcriptions", len(transcriptions))

	log.Printf("Fetching summary for meeting_id: %s", meetingID)
	summariesColl := database.Collection("summaries")
	var summary Summary
	err = summariesColl.FindOne(ctx, bson.M{"meeting_id": meetingID}).Decode(&summary)
	if err != nil {
		return nil, Summary{}, nil, fmt.Errorf("error fetching summary: %w", err)
	}
	log.Println("Fetched summary")

	log.Printf("Fetching OCR results for meeting_id: %s", meetingID)
	ocrResultsColl := database.Collection("ocr_results")
	var ocrResults []OCRResult
	cursor, err = ocrResultsColl.Find(ctx, bson.M{"meeting_id": meetingID})
	if err != nil {
		return nil, Summary{}, nil, fmt.Errorf("error fetching OCR results: %w", err)
	}
	if err = cursor.All(ctx, &ocrResults); err != nil {
		return nil, Summary{}, nil, fmt.Errorf("error decoding OCR results: %w", err)
	}
	log.Printf("Fetched %d OCR results", len(ocrResults))

	return transcriptions, summary, ocrResults, nil
}

func fetchScreenshots(meetingID string) ([]string, error) {
	screenshotsDir := fmt.Sprintf("/shared-ocr/%s", meetingID)
	files, err := ioutil.ReadDir(screenshotsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading screenshots directory: %v", err)
	}

	var screenshotPaths []string
	for _, file := range files {
		if !file.IsDir() {
			screenshotPaths = append(screenshotPaths, filepath.Join(screenshotsDir, file.Name()))
		}
	}
	return screenshotPaths, nil
}

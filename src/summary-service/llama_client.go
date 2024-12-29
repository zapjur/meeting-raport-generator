package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	apiURL  = "https://api.groq.com/openai/v1/chat/completions"
	modelID = "llama-3.1-8b-instant"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type ResponseBody struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Summary struct {
	MeetingID   string    `bson:"meeting_id"`
	SummaryText string    `bson:"summary_text"`
	CreatedAt   time.Time `bson:"created_at"`
}

func (app *Config) generateSummary() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("API key is not set.")
	}

	transcription, err := app.fetchTranscriptions()
	if err != nil {
		log.Fatalf("Error fetching transcriptions: %v", err)
	}

	prompt := `
Please provide a precise summary of the following meeting transcription:

` + transcription + `

Return only the summary without any additional text or metadata.
`

	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: prompt},
	}

	requestBody := RequestBody{
		Model:       modelID,
		Messages:    messages,
		MaxTokens:   500,
		Temperature: 0.7,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("Error serializing JSON: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatalf("Error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var response ResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error deserializing JSON: %v", err)
	}

	if len(response.Choices) > 0 {
		summaryText := response.Choices[0].Message.Content

		err = app.saveSummaryToDB("867297", summaryText)
		if err != nil {
			log.Fatalf("Error saving summary to MongoDB: %v", err)
		}

		fmt.Println("Summary saved successfully.")
	} else {
		fmt.Println("No response from model.")
	}
}

func (app *Config) saveSummaryToDB(meetingID, summaryText string) error {
	collection := app.MongoClient.Database("database").Collection("summaries")

	filter := bson.M{"meeting_id": meetingID}

	update := bson.M{
		"$set": bson.M{
			"summary_text": summaryText,
			"created_at":   time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	return err
}

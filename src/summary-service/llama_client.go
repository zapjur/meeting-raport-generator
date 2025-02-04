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

func (app *Config) generateSummary(meetingId string) error {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("API key is not set.")
	}

	transcription, err := app.fetchTranscriptions(meetingId)
	if err != nil {
		return err
	}

	prompt := `
Please provide a precise summary of the following meeting transcription:

` + transcription + `

After the summary, return a field named "statistics" containing:
- The total word count of the transcription.
- A breakdown of the word count spoken by each speaker.

Return the response in a structured format, ensuring clarity and readability.
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
		return err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response ResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	if len(response.Choices) > 0 {
		summaryText := response.Choices[0].Message.Content

		err = app.saveSummaryToDB(meetingId, summaryText)
		if err != nil {
			return err
		}

		fmt.Println("Summary saved successfully.")

		logMessage := LogMessage{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Service:   "summary",
			Level:     "INFO",
			Message:   fmt.Sprintf("Summary with meeting_id: %s saved successfully", meetingId),
			Details: map[string]interface{}{
				"meeting_id": meetingId,
				"queue":      "summary_queue",
			},
		}
		if err = app.publishLog(logMessage); err != nil {
			log.Printf("Error publishing log to RabbitMQ: %v", err)
		}
	} else {
		fmt.Println("No response from model.")

		logMessage := LogMessage{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Service:   "summary",
			Level:     "WARNING",
			Message:   fmt.Sprintf("No response from model, meeting_id: %s", meetingId),
			Details: map[string]interface{}{
				"meeting_id": meetingId,
				"queue":      "summary_queue",
			},
		}
		if err = app.publishLog(logMessage); err != nil {
			log.Printf("Error publishing log to RabbitMQ: %v", err)
		}
	}
	return nil
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

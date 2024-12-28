package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func (app *Config) generateSummary() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("API key is not set.")
	}

	transcription, err := app.fetchTranscriptions()
	if err != nil {
		log.Fatalf("Error fetching transcriptions: %v", err)
	}

	prompt := `[
		"Summarize the following transcription of the meeting:
		 ` + transcription + `",
	]`

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
		fmt.Printf("Model response: %s\n", response.Choices[0].Message.Content)
	} else {
		fmt.Println("No response from model.")
	}
}

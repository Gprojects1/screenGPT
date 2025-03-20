package proxyservicemicroservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	OpenAIAPIKey     string
	OpenAIAPIURL     string
	DefaultModel     string
	DefaultMaxTokens int
	Client           *http.Client
}

func NewController(apiKey string, apiURL string, defaultModel string, defaultMaxTokens int) *Controller {
	return &Controller{
		OpenAIAPIKey:     apiKey,
		OpenAIAPIURL:     apiURL,
		DefaultModel:     defaultModel,
		DefaultMaxTokens: defaultMaxTokens,
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (ctrl *Controller) UploadHandler(c *gin.Context) {
	var req ScreenshotRequest
	if err := c.BindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	base64Image := req.Image

	content := []Content{{
		Type: "text",
		Text: "What's in this image?",
	}, {
		Type: "image_url",
		ImageURL: ImageURL{
			URL: base64Image,
		},
	}}

	message := Message{
		Role:    "user",
		Content: content,
	}

	payload := Payload{
		Model:     ctrl.DefaultModel,
		Messages:  []Message{message},
		MaxTokens: ctrl.DefaultMaxTokens,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshaling JSON payload"})
		return
	}

	reqOpenAI, err := http.NewRequest("POST", ctrl.OpenAIAPIURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating HTTP request"})
		return
	}

	reqOpenAI.Header.Set("Content-Type", "application/json")
	reqOpenAI.Header.Set("Authorization", "Bearer "+ctrl.OpenAIAPIKey)

	client := &http.Client{}
	log.Printf("Sending payload to OpenAI: %s", string(jsonPayload))
	resp, err := client.Do(reqOpenAI)
	if err != nil {
		log.Printf("Error sending request to OpenAI: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending request to OpenAI"})
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading OpenAI response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading OpenAI response"})
		return
	}
	log.Printf("Received from OpenAI: Status Code: %d, Body: %s", resp.StatusCode, string(responseBody))
	if resp.StatusCode >= 400 {
		log.Printf("OpenAI API returned error: Status Code: %d, Body: %s", resp.StatusCode, string(responseBody))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("OpenAI API returned error: %s", string(responseBody))})
		return
	}

	var openaiResponse OpenAIResponse

	err = json.Unmarshal(responseBody, &openaiResponse)
	if err != nil {
		log.Printf("Error unmarshaling OpenAI JSON response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unmarshaling OpenAI JSON response"})
		return
	}

	var gptContent string
	if len(openaiResponse.Choices) > 0 {
		gptContent = openaiResponse.Choices[0].Message.Content
	}

	log.Println("Successfully processed OpenAI response")
	c.JSON(http.StatusOK, gin.H{"message": gptContent})
}

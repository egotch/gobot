package services

import (
	"ai-chatbot-web/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)


type AIClient struct {
	provider string
	baseURL  string
	model    string
	client	*http.Client
}

type OllamaRequest struct {
	Model   string `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream bool `json:"stream"`
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaResponse struct {
	Message OllamaMessage `json:"message"`
}

// NewAIClient initializes and returns a new AIClient based on environment variables.
func NewAIClient() *AIClient {
	provider := os.Getenv("AI_PROVIDER")
	if provider == "" {
		provider = "ollama"
	}

	baseURL := os.Getenv("OLLAMA_HOST")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

    // Ensure baseURL has http:// prefix
    if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
        baseURL = "http://" + baseURL
    }

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.1:8b"
	}

	return &AIClient{
		provider: provider,
		baseURL:  baseURL,
		model:    model,
		client:   &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SendMessage sends a message to the configured AI provider and returns the response.
func (ai *AIClient) SendMessage(messages []models.Message) (string, error) {
	switch ai.provider {
	case "ollama":
		return ai.sendOllamaMessage(messages)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", ai.provider)
	}
}

func (ai *AIClient) sendOllamaMessage(messages []models.Message) (string, error) {

	// convert internal message to ollam format
	ollamaMessages := make([]OllamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = OllamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	request := OllamaRequest{
		Model:    ai.model,
		Messages: ollamaMessages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := ai.client.Post(ai.baseURL+"/api/chat",
		"application/json",
		bytes.NewBuffer(jsonData),
		)
	if err != nil {
		return "", fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var response OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return response.Message.Content, nil
}

func (ai *AIClient) EstimateTokens(text string) int {
	// Simple estimation: 1 token per 4 characters
	return len(text) / 4
}

func (ai *AIClient) GetModel() string {
	return ai.model
}	

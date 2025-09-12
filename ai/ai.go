package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)



type OllamaRequest struct {
    Model    string        `json:"model"`
    Messages []ChatMessage `json:"messages"`
    Stream   bool          `json:"stream"`
}

type OllamaResponse struct {
    Message ChatMessage `json:"message"`
}

// For streaming responses, Ollama sends multiple JSON objects
type OllamaStreamResponse struct {
    Message ChatMessage `json:"message"`
    Done    bool        `json:"done"`
}

func (c *SmartConversation) SendToOllamaBatch() (string, error) {
    request := OllamaRequest{
        Model:    c.Model,
        Messages: c.Messages,
        Stream:   false,
    }
    
    jsonData, err := json.Marshal(request)
    if err != nil {
        return "", err
    }
    
    client := &http.Client{Timeout: 60 * time.Second}
    resp, err := client.Post(
        "http://localhost:11434/api/chat",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var response OllamaResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", err
    }
    
    // Add response to conversation
    c.AddMessage("assistant", response.Message.Content)
    
    return response.Message.Content, nil
}

func (c *SmartConversation) SendToOllamaStream() (string, error) {
    request := OllamaRequest{
        Model:    c.Model,
        Messages: c.getMessagesForAPI(),
        Stream:   true,
    }
    
    jsonData, err := json.Marshal(request)
    if err != nil {
        return "", err
    }
    
    client := &http.Client{Timeout: 300 * time.Second}
    resp, err := client.Post(
        "http://localhost:11434/api/chat",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	// process the streaming response
	for {
		var streamResp OllamaStreamResponse
		if err := decoder.Decode(&streamResp); err != nil {
			// at end of responses
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("stream decoding error: %v", err)
		}
		// Print each token as it is received
		fmt.Print(streamResp.Message.Content)
		fullResponse.WriteString(streamResp.Message.Content)

		// Small delay for a typewriter effect
		time.Sleep(10 * time.Millisecond)

		if streamResp.Done {
			break
		}
	}
	fmt.Println()

	// add response to conversation history
	response := fullResponse.String()
	c.AddMessage("assistant", response)

	return response, nil
}

// getMessagesForAPI converts messages to API format
// removes timestamp
func (c *SmartConversation) getMessagesForAPI() []ChatMessage {
	apiMessages := make([]ChatMessage, len(c.Messages))

	for i, msg := range c.Messages {
		apiMessages[i] = ChatMessage{
			Role:		msg.Role,
			Content: 	msg.Content,
		}
	}

	return apiMessages
}

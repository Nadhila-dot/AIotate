package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const OpenRouterEndpoint = "https://openrouter.ai/api/v1/chat/completions"

// OpenRouterRequest represents the request body for OpenRouter API
type OpenRouterRequest struct {
	Model    string              `json:"model"`
	Messages []OpenRouterMessage `json:"messages"`
}

// OpenRouterMessage represents a message in the conversation
type OpenRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents the response from OpenRouter API
type OpenRouterResponse struct {
	Choices []OpenRouterChoice `json:"choices"`
	Error   *OpenRouterError   `json:"error,omitempty"`
}

// OpenRouterChoice represents a choice in the response
type OpenRouterChoice struct {
	Message OpenRouterMessage `json:"message"`
}

// OpenRouterError represents an error from OpenRouter
type OpenRouterError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// GenerateWithOpenRouter generates a response using OpenRouter API
func GenerateWithOpenRouter(apiKey, model, systemPrompt, userPrompt string, cooldownSec int) (string, error) {
	// Apply cooldown if specified
	if cooldownSec > 0 {
		time.Sleep(time.Duration(cooldownSec) * time.Second)
	}

	// Build messages array
	messages := []OpenRouterMessage{}

	if systemPrompt != "" {
		messages = append(messages, OpenRouterMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	messages = append(messages, OpenRouterMessage{
		Role:    "user",
		Content: userPrompt,
	})

	// Build request body
	reqBody := OpenRouterRequest{
		Model:    model,
		Messages: messages,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", OpenRouterEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/yourusername/aiotate") // Optional
	req.Header.Set("X-Title", "AIotate")                                      // Optional

	// Make HTTP request
	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Unmarshal response
	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API error
	if openRouterResp.Error != nil {
		return "", fmt.Errorf("OpenRouter API error: %s", openRouterResp.Error.Message)
	}

	// Extract text from response
	if len(openRouterResp.Choices) > 0 {
		return openRouterResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response generated")
}

package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// GeminiRequest represents the request body for Gemini API
type GeminiRequest struct {
	Contents          []GeminiContent    `json:"contents"`
	SystemInstruction *GeminiInstruction `json:"systemInstruction,omitempty"`
}

// GeminiContent represents a content part
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a content part (text or inline data)
type GeminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *GeminiInlineData `json:"inlineData,omitempty"`
}

// GeminiInlineData represents inline file data
type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// GeminiInstruction represents the system instruction
type GeminiInstruction struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

// GeminiCandidate represents a candidate response
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// GenerateResponse generates a response using Gemini API
func GenerateResponse(apiKey, model, systemPrompt, userPrompt string, cooldownSec int) (string, error) {
	// Apply cooldown if specified
	if cooldownSec > 0 {
		time.Sleep(time.Duration(cooldownSec) * time.Second)
	}

	// Gemini API URL
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	// Build request body
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: userPrompt},
				},
			},
		},
	}
	if systemPrompt != "" {
		reqBody.SystemInstruction = &GeminiInstruction{
			Parts: []GeminiPart{
				{Text: systemPrompt},
			},
		}
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make HTTP request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		if msg := formatGeminiQuotaError(body); msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", fmt.Errorf("API error: %s", string(body))
	}

	// Unmarshal response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Extract text from response
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response generated")
}

// GenerateResponseWithAttachments generates a response using Gemini API with inline attachments when available.
func GenerateResponseWithAttachments(apiKey, model, systemPrompt, userPrompt string, attachments []Attachment, cooldownSec int) (string, error) {
	if cooldownSec > 0 {
		time.Sleep(time.Duration(cooldownSec) * time.Second)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	parts := []GeminiPart{{Text: userPrompt}}
	for _, att := range attachments {
		if att.Content == "" {
			continue
		}

		// For base64 payloads, use inlineData. Otherwise, append as text.
		if att.Encoding == "base64" && att.MimeType != "" {
			parts = append(parts, GeminiPart{
				InlineData: &GeminiInlineData{
					MimeType: att.MimeType,
					Data:     att.Content,
				},
			})
		} else {
			parts = append(parts, GeminiPart{Text: fmt.Sprintf("Attachment (%s, %s):\n%s", att.Name, att.MimeType, att.Content)})
		}
	}

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: parts,
			},
		},
	}
	if systemPrompt != "" {
		reqBody.SystemInstruction = &GeminiInstruction{
			Parts: []GeminiPart{{Text: systemPrompt}},
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		if msg := formatGeminiQuotaError(body); msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response generated")
}

func formatGeminiQuotaError(body []byte) string {
	msg := string(body)
	if !strings.Contains(msg, "RESOURCE_EXHAUSTED") && !strings.Contains(msg, "Quota exceeded") {
		return ""
	}

	retry := extractRetryDelay(msg)
	if retry != "" {
		return fmt.Sprintf("Gemini quota exceeded. Please retry in %s.", retry)
	}
	return "Gemini quota exceeded. Please try again later."
}

func extractRetryDelay(msg string) string {
	key := "\"retryDelay\": \""
	idx := strings.Index(msg, key)
	if idx == -1 {
		return ""
	}
	start := idx + len(key)
	end := strings.Index(msg[start:], "\"")
	if end == -1 {
		return ""
	}
	return msg[start : start+end]
}

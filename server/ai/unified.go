package ai

import (
	"context"
	"fmt"
	"strings"

	logg "nadhi.dev/sarvar/fun/logs"
)

// Generate generates a response using the configured AI provider with message history
func Generate(ctx context.Context, taskType TaskType, messages []Message) (string, error) {
	modelConfig, err := GetModelConfig(taskType)
	if err != nil {
		return "", fmt.Errorf("failed to get model config: %w", err)
	}

	logg.Info(fmt.Sprintf("Generating with %s (model: %s, task: %s)",
		modelConfig.Provider, modelConfig.Model, taskType))

	// Extract system and user prompts from messages
	var systemPrompt, userPrompt string
	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else if msg.Role == "user" {
			// Use the last user message as the main prompt
			userPrompt = msg.Content
		}
	}

	switch modelConfig.Provider {
	case ProviderGemini:
		resp, err := GenerateResponse(modelConfig.APIKey, modelConfig.Model, systemPrompt, userPrompt, 0)
		if err != nil && shouldFallbackToOpenRouter(err) {
			if fallback := fallbackOpenRouterConfig(taskType); fallback != nil {
				logg.Warning("Gemini quota exhausted; falling back to OpenRouter")
				return GenerateWithOpenRouter(fallback.APIKey, fallback.Model, systemPrompt, userPrompt, 0)
			}
		}
		return resp, err

	case ProviderOpenRouter:
		return GenerateWithOpenRouter(modelConfig.APIKey, modelConfig.Model, systemPrompt, userPrompt, 0)

	default:
		return "", fmt.Errorf("unsupported provider: %s", modelConfig.Provider)
	}
}

// GenerateWithAttachments generates a response with optional file attachments.
// For providers that don't support attachments, the attachments are appended to the prompt as raw text.
func GenerateWithAttachments(ctx context.Context, taskType TaskType, messages []Message, attachments []Attachment) (string, error) {
	modelConfig, err := GetModelConfig(taskType)
	if err != nil {
		return "", fmt.Errorf("failed to get model config: %w", err)
	}

	logg.Info(fmt.Sprintf("Generating with %s (model: %s, task: %s, attachments: %d)",
		modelConfig.Provider, modelConfig.Model, taskType, len(attachments)))

	// Extract system and user prompts from messages
	var systemPrompt, userPrompt string
	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else if msg.Role == "user" {
			// Use the last user message as the main prompt
			userPrompt = msg.Content
		}
	}

	switch modelConfig.Provider {
	case ProviderGemini:
		resp, err := GenerateResponseWithAttachments(modelConfig.APIKey, modelConfig.Model, systemPrompt, userPrompt, attachments, 0)
		if err != nil && shouldFallbackToOpenRouter(err) {
			if fallback := fallbackOpenRouterConfig(taskType); fallback != nil {
				logg.Warning("Gemini quota exhausted; falling back to OpenRouter")
				combined := AppendAttachmentsToPrompt(userPrompt, attachments)
				return GenerateWithOpenRouter(fallback.APIKey, fallback.Model, systemPrompt, combined, 0)
			}
		}
		return resp, err

	case ProviderOpenRouter:
		combined := AppendAttachmentsToPrompt(userPrompt, attachments)
		return GenerateWithOpenRouter(modelConfig.APIKey, modelConfig.Model, systemPrompt, combined, 0)

	default:
		return "", fmt.Errorf("unsupported provider: %s", modelConfig.Provider)
	}
}

func shouldFallbackToOpenRouter(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "RESOURCE_EXHAUSTED") ||
		strings.Contains(msg, "Quota exceeded") ||
		strings.Contains(msg, "generate_content_free_tier_requests") ||
		strings.Contains(msg, "429")
}

func fallbackOpenRouterConfig(taskType TaskType) *ModelConfig {
	aiConfig, err := GetAIConfig()
	if err != nil {
		return nil
	}
	if strings.TrimSpace(aiConfig.OpenRouterAPIKey) == "" {
		return nil
	}

	model := aiConfig.MainModel
	if taskType == TaskUtility {
		model = aiConfig.UtilityModel
	}
	if model == "" {
		if taskType == TaskUtility {
			model = "google/gemini-2.0-flash-exp:free"
		} else {
			model = "google/gemini-2.5-pro-exp-03-25:free"
		}
	}

	return &ModelConfig{
		Provider: ProviderOpenRouter,
		APIKey:   aiConfig.OpenRouterAPIKey,
		Model:    model,
	}
}

// GenerateSimple generates a response using simple system/user prompts (legacy)
func GenerateSimple(taskType TaskType, systemPrompt, userPrompt string) (string, error) {
	modelConfig, err := GetModelConfig(taskType)
	if err != nil {
		return "", fmt.Errorf("failed to get model config: %w", err)
	}

	logg.Info(fmt.Sprintf("Generating with %s (model: %s, task: %s)",
		modelConfig.Provider, modelConfig.Model, taskType))

	switch modelConfig.Provider {
	case ProviderGemini:
		return GenerateResponse(modelConfig.APIKey, modelConfig.Model, systemPrompt, userPrompt, 0)

	case ProviderOpenRouter:
		return GenerateWithOpenRouter(modelConfig.APIKey, modelConfig.Model, systemPrompt, userPrompt, 0)

	default:
		return "", fmt.Errorf("unsupported provider: %s", modelConfig.Provider)
	}
}

// GenerateWithRetry generates a response with retry logic
func GenerateWithRetry(taskType TaskType, systemPrompt, userPrompt string, maxRetries int) (string, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			logg.Warning(fmt.Sprintf("Retry attempt %d/%d", attempt+1, maxRetries))
		}

		response, err := GenerateSimple(taskType, systemPrompt, userPrompt)
		if err == nil {
			return response, nil
		}

		lastErr = err
		logg.Error(fmt.Sprintf("Generation attempt %d failed: %v", attempt+1, err))
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

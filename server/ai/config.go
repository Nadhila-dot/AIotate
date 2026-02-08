package ai

import (
	"fmt"

	"nadhi.dev/sarvar/fun/config"
)

// GetAIConfig retrieves the AI configuration from the config file
func GetAIConfig() (*AIConfig, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	aiConfig := &AIConfig{}

	// Get provider
	if provider, ok := cfg["AI_PROVIDER"].(string); ok {
		aiConfig.Provider = AIProvider(provider)
	} else {
		aiConfig.Provider = ProviderGemini // Default
	}

	// Get API keys
	if key, ok := cfg["GEMINI_API_KEY"].(string); ok {
		aiConfig.GeminiAPIKey = key
	}
	if key, ok := cfg["OPENROUTER_API_KEY"].(string); ok {
		aiConfig.OpenRouterAPIKey = key
	}

	// Get models
	if model, ok := cfg["AI_MAIN_MODEL"].(string); ok {
		aiConfig.MainModel = model
	}
	if model, ok := cfg["AI_UTILITY_MODEL"].(string); ok {
		aiConfig.UtilityModel = model
	}

	return aiConfig, nil
}

// GetModelConfig returns the appropriate model configuration for a task
func GetModelConfig(taskType TaskType) (*ModelConfig, error) {
	aiConfig, err := GetAIConfig()
	if err != nil {
		return nil, err
	}

	modelConfig := &ModelConfig{
		Provider: aiConfig.Provider,
	}

	// Select model based on task type
	switch taskType {
	case TaskLaTeXGeneration:
		modelConfig.Model = aiConfig.MainModel
	case TaskUtility:
		modelConfig.Model = aiConfig.UtilityModel
	default:
		modelConfig.Model = aiConfig.MainModel
	}

	// Select API key based on provider
	switch aiConfig.Provider {
	case ProviderGemini:
		if aiConfig.GeminiAPIKey == "" {
			return nil, fmt.Errorf("Gemini API key not configured")
		}
		modelConfig.APIKey = aiConfig.GeminiAPIKey

		// Set default Gemini models if not specified
		if modelConfig.Model == "" {
			if taskType == TaskUtility {
				modelConfig.Model = "gemini-2.0-flash-exp"
			} else {
				modelConfig.Model = "gemini-2.5-pro"
			}
		}

	case ProviderOpenRouter:
		if aiConfig.OpenRouterAPIKey == "" {
			return nil, fmt.Errorf("OpenRouter API key not configured")
		}
		modelConfig.APIKey = aiConfig.OpenRouterAPIKey

		// Set default OpenRouter models if not specified
		if modelConfig.Model == "" {
			if taskType == TaskUtility {
				modelConfig.Model = "google/gemini-2.0-flash-exp:free"
			} else {
				modelConfig.Model = "google/gemini-2.5-pro-exp-03-25:free"
			}
		}

	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", aiConfig.Provider)
	}

	return modelConfig, nil
}

// ValidateAIConfig validates the AI configuration
func ValidateAIConfig() error {
	aiConfig, err := GetAIConfig()
	if err != nil {
		return err
	}

	switch aiConfig.Provider {
	case ProviderGemini:
		if aiConfig.GeminiAPIKey == "" {
			return fmt.Errorf("Gemini API key is required when using Gemini provider")
		}
	case ProviderOpenRouter:
		if aiConfig.OpenRouterAPIKey == "" {
			return fmt.Errorf("OpenRouter API key is required when using OpenRouter provider")
		}
	default:
		return fmt.Errorf("invalid AI provider: %s (must be 'gemini' or 'openrouter')", aiConfig.Provider)
	}

	return nil
}

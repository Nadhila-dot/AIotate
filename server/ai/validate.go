package ai

import (
	"fmt"

	logg "nadhi.dev/sarvar/fun/logs"
)

// ValidateConfiguredAPI validates the AI configuration
func ValidateConfiguredAPI() error {
	logg.Info("Validating AI configuration...")

	// Validate config structure
	if err := ValidateAIConfig(); err != nil {
		logg.Error(fmt.Sprintf("AI configuration validation failed: %v", err))
		logg.Warning("Please check your set.json configuration")
		return err
	}

	// Get AI config to show what we're using
	aiConfig, err := GetAIConfig()
	if err != nil {
		return err
	}

	logg.Info(fmt.Sprintf("Using AI provider: %s", aiConfig.Provider))

	// Test with a simple generation
	systemPrompt := "You are a test assistant."
	testPrompt := "Respond with just the word 'OK' if you can read this."

	response, err := GenerateSimple(TaskUtility, systemPrompt, testPrompt)
	if err != nil {
		logg.Error(fmt.Sprintf("API validation failed: %v", err))
		return fmt.Errorf("API validation failed: %w", err)
	}

	if response == "" {
		return fmt.Errorf("API returned empty response")
	}

	logg.Success(fmt.Sprintf("AI API validated successfully (Provider: %s)", aiConfig.Provider))

	// Show configured models
	mainModel, _ := GetModelConfig(TaskLaTeXGeneration)
	utilityModel, _ := GetModelConfig(TaskUtility)

	if mainModel != nil {
		logg.Info(fmt.Sprintf("Main model (LaTeX): %s", mainModel.Model))
	}
	if utilityModel != nil {
		logg.Info(fmt.Sprintf("Utility model: %s", utilityModel.Model))
	}

	return nil
}

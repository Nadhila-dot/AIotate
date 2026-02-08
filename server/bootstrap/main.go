package bootstrap

import (
	"nadhi.dev/sarvar/fun/ai"
	logg "nadhi.dev/sarvar/fun/logs"
)

// Initialize sets up all required components for the application
func Initialize() {
	logg.Info("Initializing application...")

	// Initialize configs first
	InitConfigs()

	// Validate AI API configuration
	ValidateAI()

	logg.Success("Application initialization complete")
}

// ValidateAI validates the AI API key during boot
func ValidateAI() {
	logg.Info("Validating AI configuration...")

	if err := ai.ValidateConfiguredAPI(); err != nil {
		logg.Error("AI validation failed: " + err.Error())
		logg.Warning("The application will start, but AI features will not work until you configure a valid API key")
		logg.Warning("Please update set.json with your Gemini API key")
	}
}

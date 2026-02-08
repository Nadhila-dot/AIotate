package bootstrap

import (
	"os"

	"nadhi.dev/sarvar/fun/config"
	logg "nadhi.dev/sarvar/fun/logs"
)

// InitConfigs ensures that the set.json configuration file exists
// and has the required structure
func InitConfigs() {
	logg.Info("Checking configuration file...")

	if _, err := os.Stat(config.ConfigPath); os.IsNotExist(err) {
		logg.Warning("set.json not found, creating default config...")

		// Create a default config file
		defaultConfig := map[string]interface{}{
			"AI_PROVIDER":        "gemini",
			"GEMINI_API_KEY":     "",
			"OPENROUTER_API_KEY": "",
			"AI_MAIN_MODEL":      "",
			"AI_UTILITY_MODEL":   "",
			"MAX_SESSIONS":       2,
			"SHEET_QUEUE_DIR":    "./storage/queue_data",
		}

		if err := config.SaveConfig(defaultConfig); err != nil {
			logg.Error("Failed to create default config: " + err.Error())
			logg.Exit()
		}

		logg.Success("Default set.json created at " + config.ConfigPath)
		logg.Warning("Please configure your AI provider and API keys in set.json")
		logg.Info("Supported providers: 'gemini' or 'openrouter'")
	} else {
		logg.Success("Configuration file found")

		// Validate config structure
		cfg, err := config.GetConfig()
		if err != nil {
			logg.Error("Failed to read config: " + err.Error())
			logg.Exit()
		}

		// Check for required fields and add defaults if missing
		updated := false

		if _, ok := cfg["AI_PROVIDER"]; !ok {
			cfg["AI_PROVIDER"] = "gemini"
			updated = true
		}

		if _, ok := cfg["GEMINI_API_KEY"]; !ok {
			cfg["GEMINI_API_KEY"] = ""
			updated = true
		}

		if _, ok := cfg["OPENROUTER_API_KEY"]; !ok {
			cfg["OPENROUTER_API_KEY"] = ""
			updated = true
		}

		if _, ok := cfg["AI_MAIN_MODEL"]; !ok {
			cfg["AI_MAIN_MODEL"] = ""
			updated = true
		}

		if _, ok := cfg["AI_UTILITY_MODEL"]; !ok {
			cfg["AI_UTILITY_MODEL"] = ""
			updated = true
		}

		if _, ok := cfg["MAX_SESSIONS"]; !ok {
			cfg["MAX_SESSIONS"] = 2
			updated = true
		}

		if _, ok := cfg["SHEET_QUEUE_DIR"]; !ok {
			cfg["SHEET_QUEUE_DIR"] = "./storage/queue_data"
			updated = true
		}

		if updated {
			if err := config.SaveConfig(cfg); err != nil {
				logg.Warning("Failed to update config with defaults: " + err.Error())
			} else {
				logg.Info("Configuration updated with default values")
			}
		}
	}
}

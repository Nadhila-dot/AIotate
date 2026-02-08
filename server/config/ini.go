package config

import (
	"encoding/json"
	"errors"
	"os"
)

var (
	ErrAPIKeyNotSet  = errors.New("AI_API key is not set in configuration")
	ErrInvalidConfig = errors.New("invalid configuration format")
)

// SaveConfig saves the configuration to set.json
func SaveConfig(config map[string]interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath, data, 0644)
}

// UpdateConfigValue updates a specific configuration value
func UpdateConfigValue(key string, value interface{}) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	config[key] = value
	return SaveConfig(config)
}

// SetAPIKey sets the AI API key in the configuration
func SetAPIKey(apiKey string) error {
	return UpdateConfigValue("AI_API", apiKey)
}

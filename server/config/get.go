package config

import (
	"encoding/json"
	"os"
)

const ConfigPath = "./set.json"

// GetConfigValue retrieves a specific value from the config
func GetConfigValue(key string) interface{} {
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}

	return config[key]
}

// GetConfig retrieves the entire configuration
func GetConfig() (map[string]interface{}, error) {
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// GetAPIKey retrieves the AI API key from config
func GetAPIKey() (string, error) {
	config, err := GetConfig()
	if err != nil {
		return "", err
	}

	apiKey, ok := config["AI_API"].(string)
	if !ok || apiKey == "" {
		return "", ErrAPIKeyNotSet
	}

	return apiKey, nil
}

package config

import (
	"encoding/json"
	"log"
	"os"
	"path"
)

// Settings holds the configuration for the translation module.
type DeepSeekSettings struct {
	// APIKey is the API key for DeepSeek translation service.
	APIKey string `json:"api_key"`
	// BaseURL is the base URL for the DeepSeek API.
	BaseURL string `json:"base_url"`
}

// Settings holds the configuration for the translation module.
type Settings struct {
	// DeepSeek contains the settings for the DeepSeek translation service.
	DeepSeek DeepSeekSettings `json:"deepseek"`
}

const (
	settingsFileName = ".saat_settings.json"
)

var (
	// Options holds the global settings for the application.
	Options Settings
)

// SaveSettings saves the current settings to the user's home directory.
func SaveSettings() error {
	return saveSettings(getFilePath(settingsFileName))
}

// LoadSettings loads the settings from the user's home directory.
func LoadSettings() {
	filePath := getFilePath(settingsFileName)
	err := loadSettings(filePath)
	if err != nil {
		log.Printf("error loading settings: %v", err)
		Options = Settings{} // Default to empty settings if load fails
	}
}

func getFilePath(fileName string) string {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("error getting home directory: %v", err)
		homeDir = "."
	}
	filePath := path.Join(homeDir, fileName)
	return filePath
}

func loadSettings(filePath string) error {
	log.Printf("loading settings from %v", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &Options)
}

func saveSettings(filePath string) error {
	log.Printf("saving settings to %v", filePath)
	data, err := json.Marshal(Options)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func init() {
	LoadSettings()
}

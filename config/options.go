package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
)

// AIOptions represents the configuration options for AI services.
type AIOptions struct {
	// Vendor is the AI vendor to use (e.g., "deepseek").
	Vendor string `json:"vendor"`
	// Model is the specific model to use from the vendor (e.g., "deepseek-chat").
	Model string `json:"model"`
	// APIKey for the AI service.
	APIKey string `json:"api_key"`
}

// Options holds the application configuration options.
var Options struct {
	//DeepSeekAPIKey  (deprecated) API key for DeepSeek. Use AI.APIKey instead.
	DeepSeekAPIKey string `json:"deepseek_api_key"`
	// AI contains the configuration for AI services.
	AI AIOptions `json:"ai"`
	//TranslateAhead is the number of paragraphs to translate ahead.
	TranslateAhead int `json:"translate_ahead"`
	//AppSize is the application size factor
	AppSize int `json:"app_size"`
	//TranslationDocSize is the number of paragraphs to include in the translation context.
	TranslationDocSize int `json:"translation_doc_size"`
	//AutoProofread proofreads translated paragraphs immediately after translation
	AutoProofread bool `json:"auto_proofread"`
}

func init() {
	// Set default options
	Options.DeepSeekAPIKey = ""
	Options.TranslateAhead = 6
	Options.AppSize = 16
	Options.TranslationDocSize = 3
	Options.AutoProofread = true
}

// LoadOptions loads options from the config file, if it exists. Otherwise, defaults are used.
func LoadOptions() error {
	configFile := path.Join(ConfigDir(), "options.json")
	log.Printf("loading options file: %s", configFile)

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("options file does not exist, using defaults")
			return nil
		}
		return fmt.Errorf("failed to read options file: %v", err)
	}

	return json.Unmarshal(data, &Options)
}

// SaveOptions saves the current options to the config file.
func SaveOptions() error {
	log.Println("saving options")

	data, err := json.MarshalIndent(Options, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default options: %v", err)
	}

	configFile := path.Join(ConfigDir(), "options.json")
	log.Printf("writing options file: %s", configFile)
	return os.WriteFile(configFile, data, 0644)
}

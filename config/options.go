package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
)

// Options holds the application configuration options.
var Options struct {
	//DeepSeekAPIKey is the API key for DeepSeek service.
	DeepSeekAPIKey string `json:"deepseek_api_key"`
	//TranslateAhead is the number of paragraphs to translate ahead.
	TranslateAhead int `json:"translate_ahead"`
	//AppSize is the application size factor
	AppSize int `json:"app_size"`
	//TranslationContextParagraphs is the number of paragraphs to include in the translation context.
	TranslationContextParagraphs int `json:"translation_context_paragraphs"`
}

func init() {
	// Set default options
	Options.DeepSeekAPIKey = ""
	Options.TranslateAhead = 3
	Options.AppSize = 16
	Options.TranslationContextParagraphs = 3
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

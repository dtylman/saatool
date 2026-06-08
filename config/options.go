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
	//AIVendor is the AI vendor to use for translation (e.g. "deepseek", "gemini", "ollama").
	AIVendor string `json:"ai_vendor"`
	//AIModel is the model name to use. Leave empty to use the vendor's default.
	AIModel string `json:"ai_model"`
	//AIKKey is the API key for AI service.
	AIKey string `json:"ai_api_key"`
	//TranslateAhead is the number of paragraphs to translate ahead.
	TranslateAhead int `json:"translate_ahead"`
	//AppSize is the application size factor
	AppSize int `json:"app_size"`
	//TranslationDocSize is the number of paragraphs to include in the translation context.
	TranslationDocSize int `json:"translation_doc_size"`
	//AutoProofread proofreads translated paragraphs immediately after translation
	AutoProofread bool `json:"auto_proofread"`
	//SourceLanguage is the default source language for EPUB imports (e.g. "english")
	SourceLanguage string `json:"source_language"`
	//TargetLanguage is the default target language for EPUB imports (e.g. "hebrew")
	TargetLanguage string `json:"target_language"`
	//DarkMode enables the dark color theme
	DarkMode bool `json:"dark_mode"`
}

func init() {
	// Set default options
	Options.AIVendor = "deepseek"
	Options.AIModel = "deepseek-v4-flash"
	Options.AIKey = ""
	Options.TranslateAhead = 6
	Options.AppSize = 16
	Options.TranslationDocSize = 3
	Options.AutoProofread = true
	Options.SourceLanguage = ""
	Options.TargetLanguage = ""
	Options.DarkMode = true
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

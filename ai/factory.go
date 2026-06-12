package ai

import (
	"errors"

	"github.com/dtylman/saatool/config"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/deepseek"
	"github.com/zendev-sh/goai/provider/google"
)

// SupportedVendors lists the AI vendors that are supported by the application.
var SupportedVendors = []string{"deepseek", "google"}

// SupportedModels returns the list of supported models for a given vendor.
func SupportedModels(vendor string) []string {
	switch vendor {
	case "deepseek":
		return []string{"deepseek-v4-flash", "deepseek-v4-pro"}
	case "google":
		return []string{"gemini-3.1-pro-preview", "gemini-3.1-flash-lite"}
	default:
		return []string{}
	}
}

// GetLanguageModel returns a language model instance based on the configured AI vendor and model.
func GetLanguageModel() (provider.LanguageModel, error) {
	if config.Options.AI.Vendor == "" {
		return nil, errors.New("AI vendor is not configured")
	}
	if config.Options.AI.Model == "" {
		return nil, errors.New("AI model is not configured")
	}
	if config.Options.AI.APIKey == "" {
		return nil, errors.New("AI API key is not configured")
	}
	switch config.Options.AI.Vendor {
	case "deepseek":
		return deepseek.Chat(config.Options.AI.Model, deepseek.WithAPIKey(config.Options.AI.APIKey)), nil
	case "google":
		return google.Chat(config.Options.AI.Model, google.WithAPIKey(config.Options.AI.APIKey)), nil
	}

	return nil, errors.New("unsupported AI vendor")
}

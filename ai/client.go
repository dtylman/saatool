package ai

import (
	"fmt"

	"github.com/dtylman/goai"
	"github.com/dtylman/goai/chat"
	"github.com/dtylman/saatool/config"
)

// GetChatClient creates a chat client using global config options.
func GetChatClient() (chat.Client, error) {
	if config.Options.AIVendor == "" {
		return nil, fmt.Errorf("AI vendor is required (set via --ai-vendor or config)")
	}

	cc, err := goai.NewClient(config.Options.AIVendor, config.Options.AIModel, config.Options.AIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat client: %w", err)
	}
	return cc, nil
}

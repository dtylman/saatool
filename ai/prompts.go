package ai

import (
	"bytes"
	"fmt"
	"text/template"
)

// GetPrompt generates a prompt by filling in the provided template text with the given parameters.
func GetPrompt(text string, params map[string]string) (string, error) {
	tmpl, err := template.New("prompt").Parse(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse system prompt template: %v", err)
	}

	var promptBuf bytes.Buffer
	err = tmpl.Execute(&promptBuf, params)
	if err != nil {
		return "", fmt.Errorf("failed to execute system prompt template: %v", err)
	}
	return promptBuf.String(), nil
}

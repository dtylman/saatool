package ai

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"text/template"
)

// PromptStyle represents a translation style flavor.
type PromptStyle string

const (
	StyleStrict   PromptStyle = "strict"
	StyleAcademic PromptStyle = "academic"
	StyleLiterary PromptStyle = "literary"
	StyleArchaic  PromptStyle = "archaic"
	StyleRap      PromptStyle = "rap"
)

// PromptRole represents the message role in a chat completion.
type PromptRole string

const (
	RoleSystem PromptRole = "system"
	RoleUser   PromptRole = "user"
)

// PromptMethod represents a translation operation type.
type PromptMethod string

const (
	MethodBookDetails PromptMethod = "book_details"
	MethodTranslate   PromptMethod = "translate"
	MethodProofread   PromptMethod = "proofread"
	MethodFix         PromptMethod = "fix"
)

//go:embed prompts.json
var promptsJSON []byte

type promptEntry struct {
	Method string `json:"method"`
	System string `json:"system"`
	User   string `json:"user"`
}

type promptStyle struct {
	Style   string        `json:"style"`
	Prompts []promptEntry `json:"prompts"`
}

var promptStyles []promptStyle

func init() {
	if err := json.Unmarshal(promptsJSON, &promptStyles); err != nil {
		panic(fmt.Sprintf("failed to parse prompts.json: %v", err))
	}
}

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

// GetStyledPrompt looks up a prompt template from prompts.json by style, role,
// and method. It marshals the TranslationDocument to JSON and populates source_lang,
// target_lang, and data automatically.
func GetStyledPrompt(style PromptStyle, role PromptRole, method PromptMethod, doc *TranslationDocument) (string, error) {
	return GetStyledPromptWithParams(style, role, method, doc, nil)
}

// GetStyledPromptWithParams is like GetStyledPrompt, with optional extra
// template parameters.
func GetStyledPromptWithParams(style PromptStyle, role PromptRole, method PromptMethod, doc *TranslationDocument, extraParams map[string]string) (string, error) {
	params := make(map[string]string)
	if doc != nil {
		params["source_lang"] = doc.Source.Language
		params["target_lang"] = doc.Target.Language
		jsonData, err := json.Marshal(doc)
		if err != nil {
			return "", fmt.Errorf("failed to marshal translation document: %v", err)
		}
		params["data"] = string(jsonData)
	}
	for k, v := range extraParams {
		params[k] = v
	}

	return getPromptByMethod(style, role, method, params)
}

// StyleNames returns the list of available style names from prompts.json.
func StyleNames() []string {
	names := make([]string, len(promptStyles))
	for i, s := range promptStyles {
		names[i] = s.Style
	}
	return names
}

// GetBookDetailsPrompt looks up a book_details prompt template by style and role,
// marshaling the BookDetails struct to JSON for template rendering.
func GetBookDetailsPrompt(style PromptStyle, role PromptRole, book *BookDetails) (string, error) {
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return "", fmt.Errorf("failed to marshal book details: %v", err)
	}
	return getPromptByMethod(style, role, MethodBookDetails, map[string]string{
		"book_details": string(bookJSON),
	})
}

// getPromptByMethod is the internal lookup/render for any method with arbitrary params.
func getPromptByMethod(style PromptStyle, role PromptRole, method PromptMethod, params map[string]string) (string, error) {
	styleStr := string(style)
	methodStr := string(method)

	for _, s := range promptStyles {
		if s.Style != styleStr {
			continue
		}
		for _, p := range s.Prompts {
			if p.Method != methodStr {
				continue
			}
			var tmplText string
			switch role {
			case RoleSystem:
				tmplText = p.System
			case RoleUser:
				tmplText = p.User
			default:
				return "", fmt.Errorf("unknown role %q", role)
			}
			return GetPrompt(tmplText, params)
		}
		return "", fmt.Errorf("method %q not found in style %q", method, style)
	}
	return "", fmt.Errorf("style %q not found", style)
}

package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/dtylman/saatool/translation"
)

type Translator struct {
	client        *deepseek.Client
	inTranslation map[string]bool
	mutex         sync.Mutex
}

// NewTranslator creates a new translator with deep seek api key
func NewTranslator(apiKey string) *Translator {
	client := deepseek.NewClient(apiKey)
	if client == nil {
		log.Fatal("failed to create DeepSeek client")
	}
	return &Translator{client: client,
		inTranslation: make(map[string]bool),
		mutex:         sync.Mutex{},
	}
}

// GetBookDetails retrieves details about a book using the DeepSeek API.
func (t *Translator) GetBookDetails(ctx context.Context, book *BookDetails) (*BookDetails, error) {
	if book == nil {
		return nil, errors.New("book details cannot be nil")
	}

	bookRequest, err := json.Marshal(book)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal book to JSON: %v", err)
	}

	log.Printf("requesting book details for: %s", book.Title)
	resp, err := t.client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleSystem,
				Content: "You are a librarian.",
			},
			{
				Role: deepseek.ChatMessageRoleUser,
				Content: "Provide required in formation about the book. I need to fill in the provided JSON template. " +
					"Use the title and author fields to search for the book. Correct the existing fields and fill in missing fields. " +
					"Provide details about the main characters, genre, synopsis, and any other relevant information. " +
					"Make an effort to fill in all fields. I am most interested in the gender of the main characters, " +
					"as they are important for the translation effort." +
					"Return the information in the following JSON format: " + string(bookRequest),
			},
		},
		JSONMode: true,
	})

	if resp == nil {
		return nil, errors.New("received nil response from DeepSeek API")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %v", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices returned from chat completion")
	}

	log.Printf("DeepSeek response: %s", resp.Choices[0].Message.Content)

	var bookResponse BookDetails
	extractor := deepseek.NewJSONExtractor(nil)
	err = extractor.ExtractJSON(resp, &bookResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from response: %v", err)
	}

	return &bookResponse, nil
}

// TranslationContext represents the context for translation operations.
type TranslationContext struct {
	Source translation.Unit `json:"source"`
	Target translation.Unit `json:"target"`
}

func (t *Translator) SetTranslationInProgress(paragraphID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.inTranslation[paragraphID] = true
}

func (t *Translator) IsTranslationInProgress(paragraphID string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	_, exists := t.inTranslation[paragraphID]
	return exists
}

func (t *Translator) Translate(ctx context.Context, project translation.Project, paragraphIndex int) (string, error) {
	if t.client == nil {
		return "", errors.New("failed to create DeepSeek client")
	}

	if paragraphIndex < 0 || paragraphIndex >= len(project.Source.Paragraphs) {
		return "", fmt.Errorf("paragraph index %d out of range", paragraphIndex)
	}

	paragraphID := project.Source.Paragraphs[paragraphIndex].ID
	log.Printf("translating paragraph %d with ID %s from %s to %s", paragraphIndex, paragraphID, project.Source.Language, project.Target.Language)

	if t.IsTranslationInProgress(paragraphID) {
		log.Printf("translation for paragraph %d is already in progress", paragraphIndex)
		return "", fmt.Errorf("translation for paragraph %d is already in progress", paragraphIndex)
	}
	t.SetTranslationInProgress(paragraphID)
	bookDetails, err := json.Marshal(NewBookDetails(&project))
	if err != nil {
		return "", fmt.Errorf("failed to marshal book details: %v", err)
	}

	translationContext := &TranslationContext{
		Source: translation.Unit{
			Language:   project.Source.Language,
			Paragraphs: make([]translation.Paragraph, 0),
		},
		Target: translation.Unit{
			Language:   project.Target.Language,
			Paragraphs: make([]translation.Paragraph, 0),
		},
	}

	previousParagraphsCount := 3
	fromParagraphIndex := paragraphIndex - previousParagraphsCount
	if fromParagraphIndex < 0 {
		fromParagraphIndex = 0
	}
	for i := fromParagraphIndex; i <= paragraphIndex; i++ {
		translationContext.Source.Paragraphs = append(translationContext.Source.Paragraphs, project.Source.Paragraphs[i])
		translationContext.Target.Paragraphs = append(translationContext.Target.Paragraphs, project.Target.Paragraphs[i])
	}

	data, err := json.Marshal(translationContext)
	if err != nil {
		return "", fmt.Errorf("failed to marshal translation context: %v", err)
	}

	request := deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role: deepseek.ChatMessageRoleSystem,
				Content: fmt.Sprintf(
					"You are the perfect translator from '%v' to '%v'. You are native speaker of both %v and %v languages. "+
						"You now translating the text in the provided json: %v "+
						"Make sure to translate the text accurately and preserve its meaning and the writer style.",
					string(bookDetails),
					project.Source.Language, project.Target.Language, project.Source.Language, project.Target.Language),
			},
			{
				Role: deepseek.ChatMessageRoleUser,
				Content: "I need to provide a JSON object with translated text. The 'source' field contains a list of paragraphs in the source language, " +
					"and the 'target' field should contain the translated text in the target language. Some of them are already translated, make sure the translation is accurate, if so, keep the same ideas in the new paragraph. Keep translated names and terms consistent. " +
					"provide the translation in a JSON object. Here is the JSON object: " + string(data),
			},
		},
		JSONMode: true,
	}

	log.Printf("requesting translation for paragraph %d from %s to %s", paragraphIndex, project.Source.Language, project.Target.Language)
	resp, err := t.client.CreateChatCompletion(ctx, &request)
	if resp == nil {
		return "", errors.New("received nil response from DeepSeek API")
	}

	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no choices returned from chat completion")
	}

	log.Printf("DeepSeek translation response: %s", resp.Choices[0].Message.Content)

	var translationResponse TranslationContext
	extractor := deepseek.NewJSONExtractor(nil)
	err = extractor.ExtractJSON(resp, &translationResponse)
	if err != nil {
		return "", fmt.Errorf("failed to extract JSON from response: %v", err)
	}

	log.Printf("response: %+v", translationResponse)
	return translationResponse.Target.Paragraphs[paragraphIndex-fromParagraphIndex].Text, nil
}

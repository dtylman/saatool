package translator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
)

// BookDetails represents the details of a book.
type BookDetails struct {
	Title          string                  `json:"title"`
	Author         string                  `json:"author"`
	Synopsis       string                  `json:"synopsis"`
	Genre          string                  `json:"genre"`
	MainCharacters []translation.Character `json:"main_characters"`
}

// NewBookDetails creates a new BookDetails instance from a translation.Project.
func NewBookDetails(project *translation.Project) *BookDetails {
	return &BookDetails{
		Title:          project.Title,
		Author:         project.Author,
		Synopsis:       project.Synopsis,
		Genre:          project.Genre,
		MainCharacters: project.Characters,
	}
}

// GetBookDetails retrieves details about a book using the DeepSeek API.
func GetBookDetails(ctx context.Context, book *BookDetails) (*BookDetails, error) {

	client := deepseek.NewClient(config.Options.DeepSeek.APIKey)
	if client == nil {
		return nil, errors.New("failed to create DeepSeek client")
	}

	bookRequest, err := json.Marshal(book)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal book to JSON: %v", err)
	}

	log.Printf("requesting book details for: %s", book.Title)
	resp, err := client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
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

func Translate(ctx context.Context, project translation.Project, paragraphIndex int) (string, error) {
	client := deepseek.NewClient(config.Options.DeepSeek.APIKey)
	if client == nil {
		return "", errors.New("failed to create DeepSeek client")
	}

	if paragraphIndex < 0 || paragraphIndex >= len(project.Source.Paragraphs) {
		return "", fmt.Errorf("paragraph index %d out of range", paragraphIndex)
	}

	paragraphID := project.Source.Paragraphs[paragraphIndex].ID
	log.Printf("translating paragraph %d with ID %s from %s to %s", paragraphIndex, paragraphID, project.Source.Language, project.Target.Language)

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
	resp, err := client.CreateChatCompletion(ctx, &request)
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

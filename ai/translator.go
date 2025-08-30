package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
)

// Translator is responsible for translating text using DeepSeek API
type Translator struct {
	client        *deepseek.Client
	project       *translation.Project
	inTranslation map[string]time.Time
	mutex         sync.Mutex
	stats         *TranslationStatistics
	//OnTranslationComplete happens after a paragraph is translated and saved to the project
	OnTranslationComplete func(paragraphIndex int, translation string)
}

// NewTranslator creates a new translator with deep seek api key
func NewTranslator(project *translation.Project) (*Translator, error) {
	log.Printf("creating new translator for project: '%s'", project.GetTitle())

	client := deepseek.NewClient(config.Options.DeepSeekAPIKey)
	if client == nil {
		return nil, fmt.Errorf("failed to create DeepSeek client")
	}
	return &Translator{client: client,
		project:       project,
		inTranslation: make(map[string]time.Time),
		mutex:         sync.Mutex{},
		stats:         NewTranslationStatistics(),
	}, nil
}

// GetBookDetails retrieves details about a book using the DeepSeek API.
func (t *Translator) GetBookDetails(ctx context.Context) (*BookDetails, error) {
	book := NewBookDetails(t.project)

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

// TranslationDocument represents a specific document structure for translation requests and responses.
type TranslationDocument struct {
	// Source is the source language unit containing paragraphs to be translated.
	Source translation.Unit `json:"source"`
	// Target is the target language unit where the translated paragraphs will be stored.
	Target translation.Unit `json:"target"`
}

// SetTranslationInProgress sets the translation in progress for a given paragraph ID.
func (t *Translator) SetTranslationInProgress(paragraphID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.inTranslation[paragraphID] = time.Now()
}

// IsTranslationInProgress checks if a translation is currently in progress for a given paragraph ID.
func (t *Translator) IsTranslationInProgress(paragraphID string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	_, exists := t.inTranslation[paragraphID]
	return exists
}

// ClearTranslationInProgress clears the translation in progress for a given paragraph ID.
func (t *Translator) ClearTranslationInProgress(paragraphID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.inTranslation, paragraphID)
}

// newTranslationContext creates a new translation context for the specified paragraph index.
func (t *Translator) newTranslationContext(paragraphIndex int, sourceLang string, targetLang string, translateAhead int) *TranslationDocument {

	doc := &TranslationDocument{
		Source: translation.Unit{
			Language:   sourceLang,
			Paragraphs: make([]translation.Paragraph, 0),
		},
		Target: translation.Unit{
			Language:   targetLang,
			Paragraphs: make([]translation.Paragraph, 0),
		},
	}

	previousParagraphsCount := translateAhead
	fromParagraphIndex := paragraphIndex - previousParagraphsCount
	if fromParagraphIndex < 0 {
		fromParagraphIndex = 0
	}
	for i := fromParagraphIndex; i <= paragraphIndex; i++ {
		sourceParagraph, err := t.project.GetSourceParagraph(i)
		if err != nil {
			log.Printf("failed to get source paragraph %d: %v", i, err)
			continue
		}
		doc.Source.Paragraphs = append(doc.Source.Paragraphs, sourceParagraph)
		targetParagraph, err := t.project.GetTargetParagraph(i)
		if err != nil {
			log.Printf("failed to get target paragraph %d: %v", i, err)
			targetParagraph = translation.Paragraph{
				ID:   sourceParagraph.ID,
				Text: "",
			}
		}
		doc.Target.Paragraphs = append(doc.Target.Paragraphs, targetParagraph)
	}
	return doc
}

// FixTranslation re-translates the specified paragraph to fix its translation.
func (t *Translator) FixTranslation(ctx context.Context, paragraphIndex int) error {
	log.Printf("fixing translation for paragraph %d", paragraphIndex)
	if t.client == nil {
		return errors.New("DeepSeek client is not initialized")
	}

	sourceParagraph, err := t.project.GetSourceParagraph(paragraphIndex)
	if err != nil {
		return fmt.Errorf("failed to get source paragraph: %v", err)
	}

	sourceLang := t.project.GetSourceLanguage()
	targetLang := t.project.GetTargetLanguage()
	log.Printf("translating paragraph %d with ID %s from %s to %s", paragraphIndex, sourceParagraph.ID, sourceLang, targetLang)

	if t.IsTranslationInProgress(sourceParagraph.ID) {
		log.Printf("translation for paragraph %d is already in progress", paragraphIndex)
		return fmt.Errorf("translation for paragraph %d is already in progress", paragraphIndex)
	}
	t.SetTranslationInProgress(sourceParagraph.ID)
	defer t.ClearTranslationInProgress(sourceParagraph.ID)

	bookDetails, err := json.Marshal(NewBookDetails(t.project))
	if err != nil {
		return fmt.Errorf("failed to marshal book details: %v", err)
	}

	systemPrompt, err := GetPrompt(`You are a translation proofreader. The translation '{{.source_lang}}' to '{{.target_lang}}' was reported bad from the readers. Since you are also a native speaker of both '{{.source_lang}}' and '{{.target_lang}}', your task is to re-translate the text in the provided json. Fix any issues with translation, make an extra care to make sure all words and terms in the target language makes sense and are grammatically correct. Also make sure the target text avoids using terms from other languages. Here is some background information about the book being translated: {{.book_details}}`,
		map[string]string{
			"source_lang":  sourceLang,
			"target_lang":  targetLang,
			"book_details": string(bookDetails),
		})

	if err != nil {
		return fmt.Errorf("failed to create system prompt: %v", err)
	}

	translationContext := t.newTranslationContext(paragraphIndex, sourceLang, targetLang, 0)
	jsonData, err := json.Marshal(translationContext)
	if err != nil {
		return fmt.Errorf("failed to marshal translation context: %v", err)
	}
	userPrompt, err := GetPrompt(`The provided JSON object contains a bad 'target' translation. Please re-translate to from 'source' paragraph to '{{.target_lang}}' it and provide the corrected translation in the same JSON format. Here is the JSON object: {{.data}}`,
		map[string]string{
			"target_lang": targetLang,
			"data":        string(jsonData),
		})

	request := deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		JSONMode: true,
	}

	log.Printf("requesting fix- translation for paragraph %d from %s to %s", paragraphIndex, sourceLang, targetLang)
	resp, err := t.client.CreateChatCompletion(ctx, &request)
	if resp == nil {
		return errors.New("received nil response from DeepSeek API")
	}

	if err != nil {
		return fmt.Errorf("failed to create chat completion: %v", err)
	}

	if len(resp.Choices) == 0 {
		return errors.New("no choices returned from chat completion")
	}

	log.Printf("DeepSeek fix-translation response: %s", resp.Choices[0].Message.Content)
	var translationResponse TranslationDocument
	extractor := deepseek.NewJSONExtractor(nil)
	err = extractor.ExtractJSON(resp, &translationResponse)
	if err != nil {
		return fmt.Errorf("failed to extract JSON from response: %v", err)
	}
	log.Printf("response: %+v", translationResponse)
	translation := translationResponse.Target.Paragraphs[0].Text
	if translation == "" {
		return errors.New("received empty translation from DeepSeek API")
	}
	log.Printf("fixed translated paragraph %d: %s", paragraphIndex, translation)
	err = t.project.SetTranslation(paragraphIndex, translation)
	if err != nil {
		return fmt.Errorf("failed to set translation for paragraph %d: %v", paragraphIndex, err)
	}
	_, err = t.project.Save()
	if err != nil {
		return fmt.Errorf("failed to save project after fixing translation for paragraph %d: %v", paragraphIndex, err)
	}
	if t.OnTranslationComplete != nil {
		t.OnTranslationComplete(paragraphIndex, translation)
	}
	return nil
}

// Translate translates the specified paragraph and updates the project with the translation.
func (t *Translator) Translate(ctx context.Context, paragraphIndex int) error {
	log.Printf("translating paragraph %d", paragraphIndex)
	translation, err := t.TranslateParagraph(ctx, paragraphIndex)
	if err != nil {
		return err
	}
	err = t.project.SetTranslation(paragraphIndex, translation)
	if err != nil {
		return fmt.Errorf("failed to set translation for paragraph %d: %v", paragraphIndex, err)
	}

	if t.OnTranslationComplete != nil {
		t.OnTranslationComplete(paragraphIndex, translation)
	}

	return nil
}

// Translate translates a paragraph from the source language to the target language using the DeepSeek API and returns the translated text.
func (t *Translator) TranslateParagraph(ctx context.Context, paragraphIndex int) (string, error) {
	if t.client == nil {
		return "", errors.New("DeepSeek client is not initialized")
	}

	sourceParagraph, err := t.project.GetSourceParagraph(paragraphIndex)
	if err != nil {
		return "", fmt.Errorf("failed to get source paragraph: %v", err)
	}

	sourceLang := t.project.GetSourceLanguage()
	targetLang := t.project.GetTargetLanguage()
	log.Printf("translating paragraph %d with ID %s from %s to %s", paragraphIndex, sourceParagraph.ID, sourceLang, targetLang)

	if t.IsTranslationInProgress(sourceParagraph.ID) {
		log.Printf("translation for paragraph %d is already in progress", paragraphIndex)
		return "", fmt.Errorf("translation for paragraph %d is already in progress", paragraphIndex)
	}
	t.SetTranslationInProgress(sourceParagraph.ID)
	defer t.ClearTranslationInProgress(sourceParagraph.ID)

	t.stats.Started(sourceParagraph.ID, len(sourceParagraph.Text))
	defer t.stats.Completed(sourceParagraph.ID)

	translationContext := t.newTranslationContext(paragraphIndex, sourceLang, targetLang, config.Options.TranslateAhead)

	data, err := json.Marshal(translationContext)
	if err != nil {
		return "", fmt.Errorf("failed to marshal translation context: %v", err)
	}

	bookDetails, err := json.Marshal(NewBookDetails(t.project))
	if err != nil {
		return "", fmt.Errorf("failed to marshal book details: %v", err)
	}

	systemPrompt, err := GetPrompt(`You are a professional translator from '{{.source_lang}}' to '{{.target_lang}}' and a native speaker of both '{{.source_lang}}' and '{{.target_lang}}'. Your task is to translate '{{.book_title}}', which is a {{.book_type}}. The translation is done paragraph by paragraph. Make sure to translate the text accurately and preserve its meaning and the writer style. The translation should be: accurate; preserve the meaning and style of the original text; be free of grammatical errors; use natural and fluent {{.target_lang}} language; be culturally precise for contemporary {{.target_lang}} readers. Here are some details about the book: {{.book_details}}`,
		map[string]string{
			"source_lang":  sourceLang,
			"target_lang":  targetLang,
			"book_title":   t.project.GetTitle(),
			"book_type":    t.project.GetType(),
			"book_details": string(bookDetails),
		})

	if err != nil {
		return "", fmt.Errorf("failed to create system prompt: %v", err)
	}

	userPrompt := `I need to provide a JSON object with translated text. The 'source' field contains a list of paragraphs in the source language, and the 'target' field should contain the translated text in the target language. Some of them are already translated, make sure the translation is accurate, if so, keep the same ideas in the new paragraph. Keep translated names and terms consistent. provide the translation in a JSON object. Here is the JSON object: ` + string(data)

	request := deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		JSONMode: true,
	}

	log.Printf("requesting translation for paragraph %d from %s to %s", paragraphIndex, sourceLang, targetLang)
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

	var translationResponse TranslationDocument
	extractor := deepseek.NewJSONExtractor(nil)
	err = extractor.ExtractJSON(resp, &translationResponse)
	if err != nil {
		return "", fmt.Errorf("failed to extract JSON from response: %v", err)
	}

	log.Printf("response: %+v", translationResponse)
	translation := translationResponse.Target.Paragraphs[len(translationResponse.Target.Paragraphs)-1].Text
	if translation == "" {
		return "", errors.New("received empty translation from DeepSeek API")
	}
	log.Printf("translated paragraph %d: %s", paragraphIndex, translation)

	return translation, nil
}

// Stats returns the estimated time remaining for the next paragraph to complete and the number of paragraphs currently being translated.
func (t *Translator) Stats() (time.Duration, int) {
	return t.stats.NextETA(), t.stats.InProgressCount()
}

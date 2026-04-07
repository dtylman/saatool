package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
)

// TranslationDocument represents a specific document structure for translation requests and responses.
type TranslationDocument struct {
	// Source is the source language unit containing paragraphs to be translated.
	Source translation.Unit `json:"source"`
	// Target is the target language unit where the translated paragraphs will be stored.
	Target translation.Unit `json:"target"`
}

// Translator is responsible for translating text using DeepSeek API
type Translator struct {
	client        *deepseek.Client
	project       *translation.Project
	inTranslation map[string]time.Time
	mutex         sync.Mutex
	stats         *TranslationStatistics
	style         PromptStyle
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
		style:         StyleStrict,
	}, nil
}

// SetStyle sets the translation prompt style.
func (t *Translator) SetStyle(style PromptStyle) {
	t.style = style
}

// GetBookDetails retrieves details about a book using the DeepSeek API.
func (t *Translator) GetBookDetails(ctx context.Context) (*BookDetails, error) {
	book := NewBookDetails(t.project)

	log.Printf("requesting book details for: %s", book.Title)

	systemPrompt, err := GetBookDetailsPrompt(t.style, RoleSystem, book)
	if err != nil {
		return nil, fmt.Errorf("failed to create system prompt: %v", err)
	}
	userPrompt, err := GetBookDetailsPrompt(t.style, RoleUser, book)
	if err != nil {
		return nil, fmt.Errorf("failed to create user prompt: %v", err)
	}

	resp, err := t.client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
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
	})

	if resp == nil && err == nil {
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

// IsTranslationInProgress checks if a translation is in progress for a given paragraph ID.
func (t *Translator) SetTranslationInProgress(paragraphID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, exists := t.inTranslation[paragraphID]; exists {
		return fmt.Errorf("translation for paragraph ID %s is already in progress", paragraphID)
	}
	t.inTranslation[paragraphID] = time.Now()
	return nil
}

// ClearTranslationInProgress clears the translation in progress for a given paragraph ID.
func (t *Translator) ClearTranslationInProgress(paragraphID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.inTranslation, paragraphID)
}

// newTranslationDocument creates a new translation document for the specified paragraph index.
func (t *Translator) newTranslationDocument(paragraphIndex int, sourceLang string, targetLang string, docSize int) *TranslationDocument {

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

	previousParagraphsCount := docSize - 1
	if previousParagraphsCount < 0 {
		previousParagraphsCount = 0
	}
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

type translationRequestContext struct {
	sourceParagraph translation.Paragraph
	sourceLang      string
	targetLang      string
	paragraphIndex  int
}

func (t *Translator) newTranslationRequestContext(paragraphIndex int) (*translationRequestContext, error) {
	if t.client == nil {
		return nil, errors.New("DeepSeek client is not initialized")
	}
	sourceParagraph, err := t.project.GetSourceParagraph(paragraphIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get source paragraph: %v", err)
	}

	sourceLang := t.project.GetSourceLanguage()
	targetLang := t.project.GetTargetLanguage()
	if sourceLang == "" || targetLang == "" {
		return nil, errors.New("source or target language not set")
	}

	err = t.SetTranslationInProgress(sourceParagraph.ID)
	if err != nil {
		return nil, err
	}

	return &translationRequestContext{
		sourceParagraph: sourceParagraph,
		sourceLang:      sourceLang,
		targetLang:      targetLang,
		paragraphIndex:  paragraphIndex,
	}, nil
}

// SimpleProofRead performs a simple proofread of the specified paragraph.
func (t *Translator) SimpleProofRead(ctx context.Context, paragraphIndex int) error {
	log.Printf("proofreading paragraph %d", paragraphIndex)
	rc, err := t.newTranslationRequestContext(paragraphIndex)
	if err != nil {
		return err
	}
	defer t.ClearTranslationInProgress(rc.sourceParagraph.ID)
	t.stats.Started(rc.sourceParagraph.ID, len(rc.sourceParagraph.Text))
	defer t.stats.Completed(rc.sourceParagraph.ID)

	doc := t.newTranslationDocument(paragraphIndex, rc.sourceLang, rc.targetLang, 0)

	systemPrompt, err := GetStyledPrompt(t.style, RoleSystem, MethodProofread, doc)
	if err != nil {
		return fmt.Errorf("failed to create system prompt: %v", err)
	}

	userPrompt, err := GetStyledPrompt(t.style, RoleUser, MethodProofread, doc)
	if err != nil {
		return fmt.Errorf("failed to create user prompt: %v", err)
	}
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

	log.Printf("requesting proofreading for paragraph %d from %s to %s", paragraphIndex, rc.sourceLang, rc.targetLang)
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
	log.Printf("DeepSeek proofreading response: %s", resp.Choices[0].Message.Content)
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
	log.Printf("proofread paragraph %d: %s", paragraphIndex, translation)
	err = t.project.SetTranslation(paragraphIndex, translation)
	if err != nil {
		return fmt.Errorf("failed to set translation for paragraph %d: %v", paragraphIndex, err)
	}
	t.onTranslated(paragraphIndex, translation)
	return nil
}

// FixTranslation re-translates the specified paragraph to fix its translation.
func (t *Translator) FixTranslation(ctx context.Context, paragraphIndex int) error {
	log.Printf("fixing translation for paragraph %d", paragraphIndex)
	rc, err := t.newTranslationRequestContext(paragraphIndex)
	if err != nil {
		return err
	}
	defer t.ClearTranslationInProgress(rc.sourceParagraph.ID)
	t.stats.Started(rc.sourceParagraph.ID, len(rc.sourceParagraph.Text))
	defer t.stats.Completed(rc.sourceParagraph.ID)

	translationDocument := t.newTranslationDocument(paragraphIndex, rc.sourceLang, rc.targetLang, 0)

	systemPrompt, err := GetStyledPrompt(t.style, RoleSystem, MethodFix, translationDocument)
	if err != nil {
		return fmt.Errorf("failed to create system prompt: %v", err)
	}

	userPrompt, err := GetStyledPrompt(t.style, RoleUser, MethodFix, translationDocument)
	if err != nil {
		return fmt.Errorf("failed to create user prompt: %v", err)
	}

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

	log.Printf("requesting fix- translation for paragraph %d from %s to %s", paragraphIndex, rc.sourceLang, rc.targetLang)
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
	t.onTranslated(paragraphIndex, translation)
	return nil
}

// Translate translates the specified paragraph and updates the project with the translation.
func (t *Translator) Translate(ctx context.Context, paragraphIndex int) error {
	log.Printf("translating paragraph %d", paragraphIndex)
	err := t.TranslateParagraph(ctx, paragraphIndex)
	if err != nil {
		return err
	}
	if config.Options.AutoProofread {
		log.Printf("auto-proofreading paragraph %d", paragraphIndex)
		err = t.SimpleProofRead(ctx, paragraphIndex)
		if err != nil {
			return err
		}
	}
	return nil
}

// Translate translates a paragraph from the source language to the target language using the DeepSeek API and returns the translated text.
func (t *Translator) TranslateParagraph(ctx context.Context, paragraphIndex int) error {
	log.Printf("translating paragraph %d", paragraphIndex)

	rc, err := t.newTranslationRequestContext(paragraphIndex)
	if err != nil {
		return err
	}

	defer t.ClearTranslationInProgress(rc.sourceParagraph.ID)
	t.stats.Started(rc.sourceParagraph.ID, len(rc.sourceParagraph.Text))
	defer t.stats.Completed(rc.sourceParagraph.ID)

	translationDocument := t.newTranslationDocument(paragraphIndex, rc.sourceLang, rc.targetLang, config.Options.TranslationDocSize)

	systemPrompt, err := GetStyledPrompt(t.style, RoleSystem, MethodTranslate, translationDocument)
	if err != nil {
		return fmt.Errorf("failed to create system prompt: %v", err)
	}

	userPrompt, err := GetStyledPrompt(t.style, RoleUser, MethodTranslate, translationDocument)
	if err != nil {
		return fmt.Errorf("failed to create user prompt: %v", err)
	}

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

	log.Printf("requesting translation for paragraph %d from %s to %s", paragraphIndex, rc.sourceLang, rc.targetLang)
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

	log.Printf("DeepSeek translation response: %s", resp.Choices[0].Message.Content)

	var translationResponse TranslationDocument
	extractor := deepseek.NewJSONExtractor(nil)
	err = extractor.ExtractJSON(resp, &translationResponse)
	if err != nil {
		return fmt.Errorf("failed to extract JSON from response: %v", err)
	}

	log.Printf("response: %+v", translationResponse)
	translation := translationResponse.Target.Paragraphs[len(translationResponse.Target.Paragraphs)-1].Text
	if translation == "" {
		return errors.New("received empty translation from DeepSeek API")
	}
	log.Printf("translated paragraph %d: %s", paragraphIndex, translation)

	err = t.project.SetTranslation(paragraphIndex, translation)
	if err != nil {
		return fmt.Errorf("failed to set translation for paragraph %d: %v", paragraphIndex, err)
	}

	t.onTranslated(paragraphIndex, translation)
	return nil

}

func (t *Translator) onTranslated(paragraphIndex int, translation string) {
	if t.OnTranslationComplete != nil {
		t.OnTranslationComplete(paragraphIndex, translation)
	}
}

// Stats returns the estimated time remaining for the next paragraph to complete and the number of paragraphs currently being translated.
func (t *Translator) Stats() (time.Duration, int) {
	return t.stats.NextETA(), t.stats.InProgressCount()
}

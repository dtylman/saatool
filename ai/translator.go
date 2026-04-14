package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
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
	defaultModel  string
	fallbackModel string
	nextModel     string
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
	style := PromptStyle(project.Style)
	if style == "" {
		style = StyleStrict
	}
	defaultModel := normalizeModel(config.Options.DeepSeekModel)
	fallbackModel := strings.TrimSpace(config.Options.DeepSeekFallbackModel)
	return &Translator{client: client,
		project:       project,
		inTranslation: make(map[string]time.Time),
		mutex:         sync.Mutex{},
		stats:         NewTranslationStatistics(),
		style:         style,
		defaultModel:  defaultModel,
		fallbackModel: fallbackModel,
	}, nil
}

func normalizeModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return deepseek.DeepSeekChat
	}
	return model
}

// SetDefaultModel updates the default model used for translation calls.
func (t *Translator) SetDefaultModel(model string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.defaultModel = normalizeModel(model)
}

// SetFallbackModel updates an optional fallback model used after failed calls.
func (t *Translator) SetFallbackModel(model string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.fallbackModel = strings.TrimSpace(model)
}

// UseModelOnce schedules a one-shot model override for the next API call.
func (t *Translator) UseModelOnce(model string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.nextModel = normalizeModel(model)
}

func (t *Translator) consumeModel() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.nextModel != "" {
		model := t.nextModel
		t.nextModel = ""
		return model
	}
	return normalizeModel(t.defaultModel)
}

func (t *Translator) getFallbackModel() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return strings.TrimSpace(t.fallbackModel)
}

func (t *Translator) createChatCompletion(ctx context.Context, operation string, request *deepseek.ChatCompletionRequest) (*deepseek.ChatCompletionResponse, error) {
	if strings.TrimSpace(request.Model) == "" {
		request.Model = t.consumeModel()
	} else {
		request.Model = normalizeModel(request.Model)
	}
	resp, err := t.client.CreateChatCompletion(ctx, request)
	if err == nil && resp != nil && len(resp.Choices) > 0 {
		return resp, nil
	}

	primaryErr := validateCompletionResponse(operation, request.Model, resp, err)
	fallbackModel := t.getFallbackModel()
	if fallbackModel == "" || fallbackModel == request.Model {
		return nil, primaryErr
	}

	fallbackReq := *request
	fallbackReq.Model = fallbackModel
	log.Printf("%s failed with model '%s': %v. Retrying with fallback model '%s'", operation, request.Model, primaryErr, fallbackModel)
	fallbackResp, fallbackErr := t.client.CreateChatCompletion(ctx, &fallbackReq)
	if fallbackErr == nil && fallbackResp != nil && len(fallbackResp.Choices) > 0 {
		return fallbackResp, nil
	}

	fallbackWrappedErr := validateCompletionResponse(operation, fallbackModel, fallbackResp, fallbackErr)
	return nil, fmt.Errorf("%s failed on primary model '%s' and fallback model '%s': primary=%v fallback=%v", operation, request.Model, fallbackModel, primaryErr, fallbackWrappedErr)
}

func validateCompletionResponse(operation string, model string, resp *deepseek.ChatCompletionResponse, callErr error) error {
	if callErr != nil {
		return fmt.Errorf("%s call failed for model '%s': %w", operation, model, callErr)
	}
	if resp == nil {
		return fmt.Errorf("%s returned nil response for model '%s'", operation, model)
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("%s returned no choices for model '%s'", operation, model)
	}
	return nil
}

func (t *Translator) promptExtraParams() map[string]string {
	book := NewBookDetails(t.project)
	bookJSON, err := json.Marshal(book)
	bookDetails := ""
	if err != nil {
		log.Printf("failed to marshal book details for prompt context: %v", err)
	} else {
		bookDetails = string(bookJSON)
	}

	bookType := "book"
	if t.project.GetType() == "article" {
		bookType = "article"
	}

	return map[string]string{
		"book_title":   book.Title,
		"book_type":    bookType,
		"book_details": bookDetails,
	}
}

func extractTranslationText(doc *TranslationDocument, last bool) (string, error) {
	if doc == nil {
		return "", errors.New("translation document is nil")
	}
	if len(doc.Target.Paragraphs) == 0 {
		return "", errors.New("received translation response with empty target paragraphs")
	}
	if last {
		text := doc.Target.Paragraphs[len(doc.Target.Paragraphs)-1].Text
		if text == "" {
			return "", errors.New("received empty translation from DeepSeek API")
		}
		return text, nil
	}
	text := doc.Target.Paragraphs[0].Text
	if text == "" {
		return "", errors.New("received empty translation from DeepSeek API")
	}
	return text, nil
}

// SetStyle sets the translation prompt style and updates the project.
func (t *Translator) SetStyle(style PromptStyle) {
	t.style = style
	t.project.Style = string(style)
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

	request := deepseek.ChatCompletionRequest{
		Model: t.consumeModel(),
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

	resp, err := t.createChatCompletion(ctx, "book details", &request)
	if err != nil {
		return nil, err
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

	systemPrompt, err := GetStyledPromptWithParams(t.style, RoleSystem, MethodProofread, doc, t.promptExtraParams())
	if err != nil {
		return fmt.Errorf("failed to create system prompt: %v", err)
	}

	userPrompt, err := GetStyledPromptWithParams(t.style, RoleUser, MethodProofread, doc, t.promptExtraParams())
	if err != nil {
		return fmt.Errorf("failed to create user prompt: %v", err)
	}
	request := deepseek.ChatCompletionRequest{
		Model: t.consumeModel(),
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
	resp, err := t.createChatCompletion(ctx, "proofread", &request)
	if err != nil {
		return err
	}
	log.Printf("DeepSeek proofreading response: %s", resp.Choices[0].Message.Content)
	var translationResponse TranslationDocument
	extractor := deepseek.NewJSONExtractor(nil)
	err = extractor.ExtractJSON(resp, &translationResponse)
	if err != nil {
		return fmt.Errorf("failed to extract JSON from response: %v", err)
	}
	log.Printf("response: %+v", translationResponse)
	translation, err := extractTranslationText(&translationResponse, false)
	if err != nil {
		return err
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

	systemPrompt, err := GetStyledPromptWithParams(t.style, RoleSystem, MethodFix, translationDocument, t.promptExtraParams())
	if err != nil {
		return fmt.Errorf("failed to create system prompt: %v", err)
	}

	userPrompt, err := GetStyledPromptWithParams(t.style, RoleUser, MethodFix, translationDocument, t.promptExtraParams())
	if err != nil {
		return fmt.Errorf("failed to create user prompt: %v", err)
	}

	request := deepseek.ChatCompletionRequest{
		Model: t.consumeModel(),
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
	resp, err := t.createChatCompletion(ctx, "fix translation", &request)
	if err != nil {
		return err
	}

	log.Printf("DeepSeek fix-translation response: %s", resp.Choices[0].Message.Content)
	var translationResponse TranslationDocument
	extractor := deepseek.NewJSONExtractor(nil)
	err = extractor.ExtractJSON(resp, &translationResponse)
	if err != nil {
		return fmt.Errorf("failed to extract JSON from response: %v", err)
	}
	log.Printf("response: %+v", translationResponse)
	translation, err := extractTranslationText(&translationResponse, false)
	if err != nil {
		return err
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

	systemPrompt, err := GetStyledPromptWithParams(t.style, RoleSystem, MethodTranslate, translationDocument, t.promptExtraParams())
	if err != nil {
		return fmt.Errorf("failed to create system prompt: %v", err)
	}

	userPrompt, err := GetStyledPromptWithParams(t.style, RoleUser, MethodTranslate, translationDocument, t.promptExtraParams())
	if err != nil {
		return fmt.Errorf("failed to create user prompt: %v", err)
	}

	request := deepseek.ChatCompletionRequest{
		Model: t.consumeModel(),
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
	resp, err := t.createChatCompletion(ctx, "translation", &request)
	if err != nil {
		return err
	}

	log.Printf("DeepSeek translation response: %s", resp.Choices[0].Message.Content)

	var translationResponse TranslationDocument
	extractor := deepseek.NewJSONExtractor(nil)
	err = extractor.ExtractJSON(resp, &translationResponse)
	if err != nil {
		return fmt.Errorf("failed to extract JSON from response: %v", err)
	}

	log.Printf("response: %+v", translationResponse)
	translation, err := extractTranslationText(&translationResponse, true)
	if err != nil {
		return err
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

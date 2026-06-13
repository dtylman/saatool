package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dtylman/aitasks/tasks/translate"

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
	task          *translate.Task
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

	llm, err := GetLanguageModel()
	if err != nil {
		return nil, fmt.Errorf("failed to get language model: %v", err)
	}

	return &Translator{
		task:          translate.New(llm),
		project:       project,
		inTranslation: make(map[string]time.Time),
		mutex:         sync.Mutex{},
		stats:         NewTranslationStatistics(),
	}, nil
}

// PopulateBookDetails retrieves details about a book using the DeepSeek API.
func (t *Translator) PopulateBookDetails(ctx context.Context) (*translate.ProjectContext, error) {
	book := t.project.GetTranslationContext()
	log.Printf("requesting book details for: %s", book.Title)
	return t.task.PopulateProject(ctx, book)
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

// newTranslationRequest creates a new translation document for the specified paragraph index.
func (t *Translator) newTranslationRequest(paragraphIndex int, sourceLang string, targetLang string, historySize int) (*translate.Request, error) {
	req := translate.Request{
		ProjectContext: t.project.GetTranslationContext(),
		SourceLanguage: sourceLang,
		TargetLanguage: targetLang,
		PreviousSource: make([]string, 0),
		PreviousTarget: make([]string, 0),
		Style:          t.project.Style,
	}

	from := paragraphIndex - historySize
	for i := from; i < paragraphIndex; i++ {
		source, err := t.project.GetSourceParagraph(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get source paragraph %d: %v", i, err)
		}
		target, err := t.project.GetTargetParagraph(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get target paragraph %d: %v", i, err)
		}
		req.PreviousSource = append(req.PreviousSource, source.Text)
		req.PreviousTarget = append(req.PreviousTarget, target.Text)
	}

	sourceParagraph, err := t.project.GetSourceParagraph(paragraphIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get source paragraph %d: %v", paragraphIndex, err)
	}
	req.Text = sourceParagraph.Text

	return &req, nil
}

type translationRequestContext struct {
	sourceParagraph translation.Paragraph
	sourceLang      string
	targetLang      string
	paragraphIndex  int
}

func (t *Translator) newRequestContext(paragraphIndex int) (*translationRequestContext, error) {
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

// FixTranslation re-translates the specified paragraph to fix its translation.
func (t *Translator) FixTranslation(ctx context.Context, paragraphIndex int) error {
	log.Printf("fixing translation for paragraph %d", paragraphIndex)
	rc, err := t.newRequestContext(paragraphIndex)
	if err != nil {
		return err
	}
	defer t.ClearTranslationInProgress(rc.sourceParagraph.ID)
	t.stats.Started(rc.sourceParagraph.ID, len(rc.sourceParagraph.Text))
	defer t.stats.Completed(rc.sourceParagraph.ID)

	req, err := t.newTranslationRequest(paragraphIndex, rc.sourceLang, rc.targetLang, 0)
	if err != nil {
		return fmt.Errorf("failed to create translation document: %v", err)
	}

	response, err := t.task.Fix(ctx, req, req.Text)
	if err != nil {
		return fmt.Errorf("failed to translate: %v", err)
	}

	if response.Translation == "" {
		return errors.New("received empty translation from DeepSeek API")
	}
	err = t.project.SetTranslation(paragraphIndex, response.Translation)
	if err != nil {
		return fmt.Errorf("failed to set translation for paragraph %d: %v", paragraphIndex, err)
	}
	t.onTranslated(paragraphIndex, response.Translation)
	return nil
}

// Translate translates the specified paragraph and updates the project with the translation.
func (t *Translator) Translate(ctx context.Context, paragraphIndex int) error {
	log.Printf("translating paragraph %d", paragraphIndex)
	return t.TranslateParagraph(ctx, paragraphIndex)
}

// Translate translates a paragraph from the source language to the target language using the DeepSeek API and returns the translated text.
func (t *Translator) TranslateParagraph(ctx context.Context, paragraphIndex int) error {
	log.Printf("translating paragraph %d", paragraphIndex)

	rc, err := t.newRequestContext(paragraphIndex)
	if err != nil {
		return err
	}

	defer t.ClearTranslationInProgress(rc.sourceParagraph.ID)
	t.stats.Started(rc.sourceParagraph.ID, len(rc.sourceParagraph.Text))
	defer t.stats.Completed(rc.sourceParagraph.ID)

	req, err := t.newTranslationRequest(paragraphIndex, rc.sourceLang, rc.targetLang, config.Options.TranslationDocSize)
	if err != nil {
		return fmt.Errorf("failed to create translation document: %v", err)
	}

	log.Printf("requesting translation for paragraph %d from %s to %s", paragraphIndex, rc.sourceLang, rc.targetLang)

	response, err := t.task.Translate(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to translate: %v", err)
	}

	log.Printf("response: %+v", response)

	log.Printf("translated paragraph %d: %s", paragraphIndex, response.Translation)

	err = t.project.SetTranslation(paragraphIndex, response.Translation)
	if err != nil {
		return fmt.Errorf("failed to set translation for paragraph %d: %v", paragraphIndex, err)
	}

	t.onTranslated(paragraphIndex, response.Translation)
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

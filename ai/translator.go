package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dtylman/goai/chat"
	"github.com/dtylman/goai/tasks/translate"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
)

// Translator orchestrates paragraph translation using goai.
type Translator struct {
	client        chat.Client
	task          *translate.Task
	project       *translation.Project
	inTranslation map[string]time.Time
	mutex         sync.Mutex
	stats         *TranslationStatistics
	// OnTranslationComplete happens after a paragraph is translated and saved to the project
	OnTranslationComplete func(paragraphIndex int, translation string)
}

// NewTranslator creates a new Translator for the given project.
func NewTranslator(project *translation.Project) (*Translator, error) {
	log.Printf("creating new translator for project: '%s'", project.GetTitle())

	client, err := GetChatClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create chat client: %w", err)
	}
	task := translate.New(client)
	task.AutoProofread = config.Options.AutoProofread

	return &Translator{
		client:        client,
		task:          task,
		project:       project,
		inTranslation: make(map[string]time.Time),
		stats:         NewTranslationStatistics(),
	}, nil
}

// ── In-flight deduplication ───────────────────────────────────────────────────

// SetTranslationInProgress marks a paragraph as currently being translated.
// Returns an error if that paragraph ID is already in-flight.
func (t *Translator) SetTranslationInProgress(paragraphID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, exists := t.inTranslation[paragraphID]; exists {
		return fmt.Errorf("translation for paragraph ID %s is already in progress", paragraphID)
	}
	t.inTranslation[paragraphID] = time.Now()
	return nil
}

// ClearTranslationInProgress removes the in-progress marker for a paragraph.
func (t *Translator) ClearTranslationInProgress(paragraphID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.inTranslation, paragraphID)
}

// ── translationRequestContext ─────────────────────────────────────────────────

type translationRequestContext struct {
	sourceParagraph translation.Paragraph
	sourceLang      string
	targetLang      string
	paragraphIndex  int
}

func (t *Translator) newTranslationRequestContext(paragraphIndex int) (*translationRequestContext, error) {
	if t.client == nil {
		return nil, errors.New("chat client is not initialized")
	}
	src, err := t.project.GetSourceParagraph(paragraphIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get source paragraph: %v", err)
	}

	sourceLang := t.project.GetSourceLanguage()
	targetLang := t.project.GetTargetLanguage()
	if sourceLang == "" || targetLang == "" {
		return nil, errors.New("source or target language not set")
	}

	if err := t.SetTranslationInProgress(src.ID); err != nil {
		return nil, err
	}

	return &translationRequestContext{
		sourceParagraph: src,
		sourceLang:      sourceLang,
		targetLang:      targetLang,
		paragraphIndex:  paragraphIndex,
	}, nil
}

func (t *Translator) collectContextParagraphs(paragraphIndex int, maxContext int) ([]string, []string) {
	if maxContext <= 0 {
		return nil, nil
	}

	indices := make([]int, 0, maxContext)
	for i := paragraphIndex - 1; i >= 0 && len(indices) < maxContext; i-- {
		tgt, err := t.project.GetTargetParagraph(i)
		if err == nil && strings.TrimSpace(tgt.Text) != "" {
			indices = append(indices, i)
		}
	}
	for i, j := 0, len(indices)-1; i < j; i, j = i+1, j-1 {
		indices[i], indices[j] = indices[j], indices[i]
	}

	prevSource := make([]string, 0, len(indices))
	prevTarget := make([]string, 0, len(indices))
	for _, idx := range indices {
		src, err := t.project.GetSourceParagraph(idx)
		if err != nil {
			continue
		}
		tgt, err := t.project.GetTargetParagraph(idx)
		if err != nil {
			continue
		}
		prevSource = append(prevSource, src.Text)
		prevTarget = append(prevTarget, tgt.Text)
	}

	return prevSource, prevTarget
}

func (t *Translator) newTranslateRequest(paragraphIndex int) (*translationRequestContext, *translate.Request, error) {
	rc, err := t.newTranslationRequestContext(paragraphIndex)
	if err != nil {
		return nil, nil, err
	}

	projectContext := t.project.ToProjectContext()
	req := &translate.Request{
		ProjectContext: projectContext,
		SourceLanguage: rc.sourceLang,
		TargetLanguage: rc.targetLang,
		Text:           rc.sourceParagraph.Text,
		Style:          t.project.WritingStyle,
	}

	prevSource, prevTarget := t.collectContextParagraphs(paragraphIndex, config.Options.TranslationDocSize-1)
	req.PreviousSource = prevSource
	req.PreviousTarget = prevTarget

	return rc, req, nil
}

// ── GetBookDetails ────────────────────────────────────────────────────────────

// GetBookDetails uses the AI to enrich the project's book metadata via goai PopulateProject.
func (t *Translator) GetBookDetails(ctx context.Context) error {
	log.Printf("requesting book details for: %s", t.project.GetTitle())

	populated, err := t.task.PopulateProject(ctx, t.project.ToProjectContext())
	if err != nil {
		return fmt.Errorf("failed to populate project: %w", err)
	}
	t.project.PopulateFromProjectContext(populated)
	return nil
}

// ── SimpleProofRead ───────────────────────────────────────────────────────────

// SimpleProofRead proofreads the translated text of the specified paragraph.
func (t *Translator) SimpleProofRead(ctx context.Context, paragraphIndex int) error {
	log.Printf("proofreading paragraph %d", paragraphIndex)
	rc, req, err := t.newTranslateRequest(paragraphIndex)
	if err != nil {
		return err
	}
	defer t.ClearTranslationInProgress(rc.sourceParagraph.ID)
	t.stats.Started(rc.sourceParagraph.ID, len(rc.sourceParagraph.Text))
	defer t.stats.Completed(rc.sourceParagraph.ID)

	target, err := t.project.GetTargetParagraph(paragraphIndex)
	if err != nil {
		return fmt.Errorf("failed to get target paragraph %d: %v", paragraphIndex, err)
	}
	draft := strings.TrimSpace(target.Text)
	if draft == "" {
		return errors.New("no target text available for proofreading")
	}

	result, err := t.task.Proofread(ctx, req, draft)
	if err != nil {
		return fmt.Errorf("proofread failed for paragraph %d: %w", paragraphIndex, err)
	}

	proofed := strings.TrimSpace(result.Translation)
	if proofed == "" {
		return errors.New("received empty proofread text")
	}
	if err := t.project.SetTranslation(paragraphIndex, proofed); err != nil {
		return fmt.Errorf("failed to set translation for paragraph %d: %v", paragraphIndex, err)
	}
	t.onTranslated(paragraphIndex, proofed)
	return nil
}

// ── FixTranslation ────────────────────────────────────────────────────────────

// FixTranslation re-translates the specified paragraph to correct a bad translation.
func (t *Translator) FixTranslation(ctx context.Context, paragraphIndex int) error {
	log.Printf("fixing translation for paragraph %d", paragraphIndex)
	rc, req, err := t.newTranslateRequest(paragraphIndex)
	if err != nil {
		return err
	}
	defer t.ClearTranslationInProgress(rc.sourceParagraph.ID)
	t.stats.Started(rc.sourceParagraph.ID, len(rc.sourceParagraph.Text))
	defer t.stats.Completed(rc.sourceParagraph.ID)

	target, err := t.project.GetTargetParagraph(paragraphIndex)
	if err != nil {
		return fmt.Errorf("failed to get target paragraph %d: %v", paragraphIndex, err)
	}
	badTranslation := strings.TrimSpace(target.Text)
	if badTranslation == "" {
		return errors.New("no target text available to fix")
	}

	result, err := t.task.Fix(ctx, req, badTranslation)
	if err != nil {
		return fmt.Errorf("fix translation failed for paragraph %d: %w", paragraphIndex, err)
	}

	fixed := strings.TrimSpace(result.Translation)
	if fixed == "" {
		return errors.New("received empty fixed translation")
	}
	if err := t.project.SetTranslation(paragraphIndex, fixed); err != nil {
		return fmt.Errorf("failed to set translation for paragraph %d: %v", paragraphIndex, err)
	}
	t.onTranslated(paragraphIndex, fixed)
	return nil
}

// ── Translate / TranslateParagraph ────────────────────────────────────────────

// Translate translates the paragraph and optionally auto-proofreads it.
func (t *Translator) Translate(ctx context.Context, paragraphIndex int) error {
	log.Printf("translating paragraph %d", paragraphIndex)
	if err := t.TranslateParagraph(ctx, paragraphIndex); err != nil {
		return err
	}
	if config.Options.AutoProofread {
		log.Printf("auto-proofreading paragraph %d", paragraphIndex)
		if err := t.SimpleProofRead(ctx, paragraphIndex); err != nil {
			return err
		}
	}
	return nil
}

// TranslateParagraph sends a paragraph to the DeepSeek API for translation.
func (t *Translator) TranslateParagraph(ctx context.Context, paragraphIndex int) error {
	log.Printf("translating paragraph %d", paragraphIndex)

	// ── Change 3: skip if already translated ─────────────────────────────────
	if existing, err := t.project.GetTargetParagraph(paragraphIndex); err == nil && existing.Text != "" {
		log.Printf("paragraph %d already translated — skipping API call", paragraphIndex)
		t.onTranslated(paragraphIndex, existing.Text)
		return nil
	}

	rc, err := t.newTranslationRequestContext(paragraphIndex)
	rc, req, err := t.newTranslateRequest(paragraphIndex)
	if err != nil {
		return err
	}
	defer t.ClearTranslationInProgress(rc.sourceParagraph.ID)
	t.stats.Started(rc.sourceParagraph.ID, len(rc.sourceParagraph.Text))
	defer t.stats.Completed(rc.sourceParagraph.ID)

	result, err := t.task.Translate(ctx, req)
	if err != nil {
		return fmt.Errorf("translate failed for paragraph %d: %w", paragraphIndex, err)
	}

	translated := strings.TrimSpace(result.Translation)
	if translated == "" {
		return errors.New("received empty translation")
	}

	if err := t.project.SetTranslation(paragraphIndex, translated); err != nil {
		return fmt.Errorf("failed to set translation for paragraph %d: %v", paragraphIndex, err)
	}
	t.onTranslated(paragraphIndex, translated)
	return nil
}

func (t *Translator) onTranslated(paragraphIndex int, translation string) {
	if t.OnTranslationComplete != nil {
		t.OnTranslationComplete(paragraphIndex, translation)
	}
}

// Stats returns the estimated time until the next paragraph completes and
// the number of paragraphs currently being translated.
func (t *Translator) Stats() (time.Duration, int) {
	return t.stats.NextETA(), t.stats.InProgressCount()
}

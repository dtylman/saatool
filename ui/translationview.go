package ui

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
	"github.com/dustin/go-humanize"
)

// TranslationView represents the view for translating text in a project.
type TranslationView struct {
	View           fyne.CanvasObject
	project        *translation.Project
	translator     *ai.Translator
	txt            *widgets.BidiText
	panelMain      *fyne.Container
	btnProgress    *widget.Button
	sourceView     bool   // true for source language, false for target language
	paragraphIndex int    // current paragraph index
	projectHash    []byte // hash of the project to detect changes

}

// NewTranslationView creates a new TranslationView for the given project.
func NewTranslationView(project *translation.Project) (*TranslationView, error) {
	translator, err := ai.NewTranslator(project)
	if err != nil {
		return nil, fmt.Errorf("failed to create translator: %v", err)
	}

	tv := &TranslationView{
		project:        project,
		txt:            widgets.NewBidiText(),
		btnProgress:    widget.NewButton("", nil),
		sourceView:     project.LastSourceView,
		paragraphIndex: project.LastParagraphIndex,
		translator:     translator,
	}

	tv.translator.OnTranslationComplete = tv.onTranslationCompleted

	tv.btnProgress = widget.NewButton("Go to Paragraph", tv.onProgressTapped)
	Main.ClearActions()
	Main.AddActionWidget(widget.NewCheck("Source", tv.onSourceChange))
	Main.AddAction("Next", widgets.IconNext, tv.onNext)
	Main.AddAction("Previous", widgets.IconPrev, tv.onPrevious)

	Main.AddActionWidget(tv.btnProgress)

	tv.txt.Direction = widgets.RightToLeft
	appSize := config.Options.AppSize
	tv.txt.TextSize = float32(appSize) * 2
	tv.txt.Padding = float32(appSize) / 2
	tv.txt.Spacing = float32(appSize) / 2

	tv.panelMain = container.NewStack(tv.txt)
	view := widgets.NewPanel(tv.panelMain, fyne.NewSize(0, 0))
	view.OnTapped = tv.onMainPanelTapped
	tv.View = view
	tv.updateProgress()
	tv.updateText()

	return tv, nil
}

// onMainPanelTapped handles tap events on the main panel of the translation view.
func (tv *TranslationView) onMainPanelTapped(pe *fyne.PointEvent) {
	leftSide := pe.Position.X < tv.txt.Size().Width/2
	ltr := tv.txt.Direction == widgets.LeftToRight

	if (leftSide && ltr) || (!leftSide && !ltr) {
		tv.onPrevious()
	} else {
		tv.onNext()
	}
}

// onNext handles the action of moving to the next word or paragraph.
func (tv *TranslationView) onNext() {
	// Move to the next word
	if tv.txt.Next() {
		tv.updateProgress()
		return
	}
	// Load the next paragraph
	tv.SetParagraph(tv.paragraphIndex + 1)
}

// SetParagraph sets the current paragraph to display in the translation view.
func (tv *TranslationView) SetParagraph(paragraph int) {
	if (tv.paragraphIndex == paragraph) || (paragraph < 0) || (paragraph >= len(tv.project.Target.Paragraphs)) {
		return
	}

	tv.paragraphIndex = paragraph
	if !tv.sourceView {
		tv.invokeTranslation()
	}

	tv.updateText()
	tv.updateProgress()
}

// invokeTranslation initiates the translation for the current and subsequent paragraphs based on the configuration.
func (tv *TranslationView) invokeTranslation() {
	err := tv.translateParagraph(tv.paragraphIndex)
	if err != nil {
		log.Printf("translation error (paragraph %v): %v", tv.paragraphIndex, err)
	}
	// Pre-translate ahead paragraphs
	for i := 1; i <= config.Options.TranslateAhead; i++ {
		idx := tv.paragraphIndex + i
		if idx >= len(tv.project.Target.Paragraphs) {
			break
		}
		err := tv.translateParagraph(idx)
		if err != nil {
			log.Printf("pre-translation error (paragraph %v): %v", idx, err)
		}
	}
}

// translateParagraph translates the specified paragraph from source to target language.
func (tv *TranslationView) translateParagraph(paragraph int) error {

	if paragraph < 0 || paragraph >= len(tv.project.Target.Paragraphs) {
		return fmt.Errorf("paragraph %d out of range", paragraph)

	}

	sourceLang := tv.project.Source.Language
	targetLang := tv.project.Target.Language

	if sourceLang == "" || targetLang == "" {
		return errors.New("source or target language not set")
	}

	log.Printf("translating paragraph %d from %v to %v", paragraph, sourceLang, targetLang)

	target := tv.project.Target.Paragraphs[paragraph]

	if target.Text != "" {
		log.Printf("paragraph %d already translated", paragraph)
		return nil
	}

	if paragraph < 0 || paragraph >= len(tv.project.Source.Paragraphs) {
		log.Printf("source paragraph %d out of range", paragraph)
		return errors.New("source paragraph out of range")
	}

	go func() {
		err := tv.translator.Translate(context.Background(), paragraph)
		if err != nil {
			log.Printf("translation error (paragraph %v): %v", paragraph, err)
		}
	}()

	return nil
}

// onTranslationCompleted is called when a translation is completed.
func (tv *TranslationView) onTranslationCompleted(paragraphIndex int, translation string) {
	if tv.paragraphIndex == paragraphIndex {
		fyne.Do(func() {
			tv.updateText()
		})
	}
}

// onProgressTapped handles the action of navigating to a specific paragraph.
func (tv *TranslationView) onProgressTapped() {
	selected := binding.NewString()
	selected.Set(fmt.Sprintf("%d", tv.paragraphIndex))
	dialog.NewForm("Go to Paragraph", "Go", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Go to:", widget.NewEntryWithData(selected)),
		},
		func(ok bool) {
			if ok {
				data, err := selected.Get()
				if err != nil {
					dialog.ShowError(fmt.Errorf("invalid paragraph number"), Main.window)
					return
				}
				selectedParagraph, err := strconv.Atoi(data)
				if err != nil {
					dialog.ShowError(fmt.Errorf("invalid paragraph number: %v", err), Main.window)
					return
				}
				// Validate the paragraph number
				if selectedParagraph < 0 || selectedParagraph >= len(tv.project.Target.Paragraphs) {
					dialog.ShowError(fmt.Errorf("paragraph number out of range"), Main.window)
					return
				}
				tv.SetParagraph(selectedParagraph)
			}
		},
		Main.window,
	).Show()
}

// onPrevious handles the action of moving to the previous word or paragraph.
func (tv *TranslationView) onPrevious() {
	// Move to the previous word
	if tv.txt.Previous() {
		tv.updateProgress()
		return
	}

	// Load the previous paragraph
	tv.SetParagraph(tv.paragraphIndex - 1)
}

// updateProgress updates the progress label with the current paragraph and word offset.
func (tv *TranslationView) updateProgress() {
	tv.btnProgress.SetText(fmt.Sprintf("p: %v.%v/%v", tv.paragraphIndex, tv.txt.Offset, len(tv.project.Target.Paragraphs)))
	tv.project.LastSourceView = tv.sourceView
	tv.project.LastParagraphIndex = tv.paragraphIndex
}

// updateText updates the text displayed in the translation view based on the current paragraph and language.
func (tv *TranslationView) updateText() {
	var p translation.Paragraph
	var lang string
	if tv.sourceView {
		lang = tv.project.Source.Language
		p = tv.project.Source.Paragraphs[tv.paragraphIndex]
	} else {
		lang = tv.project.Target.Language
		p = tv.project.Target.Paragraphs[tv.paragraphIndex]
	}

	if p.Text == "" {
		tv.panelMain.RemoveAll()
		text := "No text available for this paragraph."
		if tv.translator != nil {
			startTime := tv.translator.TranslationTime(p.ID)
			if startTime != (time.Time{}) {
				text = fmt.Sprintf("Translation in progress for paragraph %v %v", tv.paragraphIndex, humanize.Time(startTime))
			}
		}

		label := widget.NewLabel(text)
		label.Alignment = fyne.TextAlignCenter
		label.Wrapping = fyne.TextWrapBreak

		tv.panelMain.Add(label)

	} else {
		tv.panelMain.RemoveAll()
		tv.panelMain.Add(tv.txt)
		words := strings.Fields(strings.Replace(p.Text, "\n", " <NL> ", -1))
		dir := translation.GetTextDirection(lang)
		tv.txt.SetWords(words)
		if dir == translation.RightToLeft {
			tv.txt.Direction = widgets.RightToLeft
		} else {
			tv.txt.Direction = widgets.LeftToRight
		}
	}

	tv.updateProgress()

}

// onSourceChange handles the change of the source language toggle.
func (tv *TranslationView) onSourceChange(checked bool) {
	tv.sourceView = checked
	tv.updateText()
}

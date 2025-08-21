package ui

import (
	"context"
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// TranslationView represents the view for translating text in a project.
type TranslationView struct {
	View        fyne.CanvasObject
	project     *translation.Project
	txt         *widgets.BidiText
	panelMain   *fyne.Container
	lblProgress *widget.Label
	source      bool // true for source language, false for target language
	paragraph   int  // current paragraph index
}

func NewTranslationView(project *translation.Project) *TranslationView {
	prefs := Main.Preferences()

	tv := &TranslationView{
		project:     project,
		txt:         widgets.NewBidiText(),
		lblProgress: widget.NewLabel(""),
		source:      prefs.LastTranslationSource(),
		paragraph:   prefs.LastTranslationParagraph(),
	}

	Main.ClearActions()
	Main.AddActionWidget(widget.NewCheck("Source", tv.onSourceChange))
	Main.AddAction("Next", widgets.IconNext, tv.onNext)
	Main.AddAction("Previous", widgets.IconPrev, tv.onPrevious)
	Main.AddActionWidget(tv.lblProgress)
	Main.AddActionWidget(widget.NewSeparator())

	tv.txt.Direction = widgets.RightToLeft
	appSize := Main.Preferences().AppSize()
	tv.txt.TextSize = float32(appSize) * 2
	tv.txt.Padding = float32(appSize) / 2
	tv.txt.Spacing = float32(appSize) / 2

	tv.panelMain = container.NewStack(tv.txt)
	view := widgets.NewPanel(tv.panelMain, fyne.NewSize(0, 0))
	view.OnTapped = tv.onMainPanelTapped
	tv.View = view

	tv.updateProgress()
	tv.updateText()

	return tv
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
	tv.SetParagraph(tv.paragraph + 1)
}

// SetParagraph sets the current paragraph to display in the translation view.
func (tv *TranslationView) SetParagraph(paragraph int) {
	if (tv.paragraph == paragraph) || (paragraph < 0) || (paragraph >= len(tv.project.Target.Paragraphs)) {
		return
	}

	tv.paragraph = paragraph
	if !tv.source {
		tv.translate(tv.paragraph, tv.project.Source.Language, tv.project.Target.Language, false)
		translateAhead := 3
		for i := 1; i <= translateAhead; i++ {
			if tv.paragraph+i < len(tv.project.Target.Paragraphs) {
				tv.translate(tv.paragraph+i, tv.project.Source.Language, tv.project.Target.Language, false)
			}
		}
	}

	tv.updateText()
	tv.updateProgress()
}

// onPrevious handles the action of moving to the previous word or paragraph.
func (tv *TranslationView) onPrevious() {
	// Move to the previous word
	if tv.txt.Previous() {
		tv.updateProgress()
		return
	}

	// Load the previous paragraph
	tv.SetParagraph(tv.paragraph - 1)
}

// updateProgress updates the progress label with the current paragraph and word offset.
func (tv *TranslationView) updateProgress() {
	tv.lblProgress.SetText(fmt.Sprintf("p: %v.%v/%v", tv.paragraph, tv.txt.Offset, len(tv.project.Target.Paragraphs)))
	prefs := Main.Preferences()
	prefs.SetLastTranslationParagraph(tv.paragraph)
	prefs.SetLastTranslationSource(tv.source)
}

// updateText updates the text displayed in the translation view based on the current paragraph and language.
func (tv *TranslationView) updateText() {
	var p translation.Paragraph
	var lang string
	if tv.source {
		lang = tv.project.Source.Language
		p = tv.project.Source.Paragraphs[tv.paragraph]
	} else {
		lang = tv.project.Target.Language
		p = tv.project.Target.Paragraphs[tv.paragraph]
	}

	if p.Text == "" {
		tv.panelMain.RemoveAll()
		tv.panelMain.Add(widget.NewLabel("No text available for this paragraph."))

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
	tv.source = checked
	tv.updateText()
}

// translate translates the specified paragraph from source to target language.
func (tv *TranslationView) translate(paragraph int, sourceLang string, targetLang string, force bool) {
	log.Printf("translating paragraph %d from %v to %v (force=%v)", paragraph, sourceLang, targetLang, force)

	if paragraph < 0 || paragraph >= len(tv.project.Target.Paragraphs) {
		log.Printf("target paragraph %d out of range", paragraph)
		return
	}

	target := tv.project.Target.Paragraphs[paragraph]

	if target.Text != "" && !force {
		log.Printf("paragraph %d already translated", paragraph)
		return
	}

	if paragraph < 0 || paragraph >= len(tv.project.Source.Paragraphs) {
		log.Printf("source paragraph %d out of range", paragraph)
		return
	}

	go func() {
		ctx := context.Background()
		translated, err := Main.Translator().Translate(ctx, *tv.project, paragraph)
		if err != nil {
			log.Printf("translation error: %v", err)
		}
		log.Printf("translation result: %v", translated)

		if translated != "" {
			fyne.Do(func() {
				tv.project.Target.Paragraphs[paragraph].Text = translated
				tv.project.Target.Paragraphs[paragraph].ID = tv.project.Source.Paragraphs[paragraph].ID
				log.Printf("updated target paragraph %d with translation", paragraph)

				activeProject := Main.Preferences().ActiveProject()
				err = tv.project.Save(activeProject)
				if err != nil {
					log.Printf("failed to save project: %v", err)
				}
				// Update the text view if the current paragraph is being displayed
				if tv.paragraph == paragraph {
					tv.updateText()
				}
			})
		} else {
			log.Printf("no translation received for paragraph %d", paragraph)
		}
	}()
	// translate the paragraph
	// when translation is finished, update the project.
	// if the paragraph is being displayed, refresh the display.( maybe always refresh? )
}

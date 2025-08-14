package ui

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

type TranslationView struct {
	OnClose     func()
	View        fyne.CanvasObject
	project     *translation.Project
	txt         *widgets.BidiText
	lblProgress *widget.Label
	source      bool // true for source language, false for target language
	paragraph   int  // current paragraph index
}

func NewTranslationView(project *translation.Project) *TranslationView {
	tv := &TranslationView{
		project:     project,
		txt:         widgets.NewBidiText(),
		lblProgress: widget.NewLabel(""),
		source:      false, // default to target language
		paragraph:   0,     // start with the first paragraph
	}

	toolBar := container.NewHBox(
		widget.NewCheck("Source", tv.onSourceChange),
		widget.NewButton("Next", tv.onNext),
		widget.NewButton("Previous", tv.onPrevious),
		tv.lblProgress,
		widget.NewSeparator(),
		widget.NewButton("Back", tv.onClose),
	)

	tv.txt.Direction = widgets.RightToLeft
	tv.txt.TextSize = 40
	tv.txt.Padding = 10
	tv.txt.Spacing = 15

	content := container.NewBorder(
		nil,
		toolBar,
		nil,
		nil,
		tv.txt,
	)

	tv.View = content

	tv.updateText()

	return tv
}

func (tv *TranslationView) onClose() {
	if tv.OnClose != nil {
		tv.OnClose()
	}
}

func (tv *TranslationView) onNext() {
	// Move to the next word
	if tv.txt.Next() {
		tv.updateProgress()
		return
	}
	// Load the next paragraph
	if tv.paragraph < len(tv.project.Target.Paragraphs)-1 {
		tv.paragraph++
		tv.updateText()
		tv.updateProgress()
	}
}

func (tv *TranslationView) onPrevious() {
	// Move to the previous word
	if tv.txt.Previous() {
		tv.updateProgress()
		return
	}
	// Load the previous paragraph
	if tv.paragraph > 0 {
		tv.paragraph--
		tv.updateText()
		tv.updateProgress()
	}
}

func (tv *TranslationView) updateProgress() {
	tv.lblProgress.SetText(fmt.Sprintf("p: %v (%v-%v)", tv.paragraph, tv.txt.Offset, tv.txt.Length))
}

func (tv *TranslationView) updateText() {
	var text string
	if tv.source {
		tv.txt.Direction = widgets.LeftToRight
		log.Println("Fix me!")
		text = tv.project.Source.Paragraphs[tv.paragraph].Text
	} else {
		tv.txt.Direction = widgets.RightToLeft
		text = tv.project.Target.Paragraphs[tv.paragraph].Text
	}
	words := strings.Fields(strings.Replace(text, "\n", " <NL> ", -1))
	tv.txt.SetWords(words)
	tv.updateProgress()
}

func (tv *TranslationView) onSourceChange(checked bool) {
	tv.source = checked
	tv.updateText()
}

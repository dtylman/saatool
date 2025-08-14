package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

type TranslationView struct {
	OnClose   func()
	View      fyne.CanvasObject
	project   *translation.Project
	txt       *widgets.BidiText
	source    bool // true for source language, false for target language
	paragraph int  // current paragraph index
	offset    int  // current offset in the paragraph (this the word index)
}

func NewTranslationView(project *translation.Project) *TranslationView {
	tv := &TranslationView{
		project:   project,
		source:    true, // default to source language
		paragraph: 0,    // start with the first paragraph
		offset:    0,    // start at the beginning of the paragraph
	}

	toolBar := container.NewHBox(
		widget.NewLabel("Translation View"),
		widget.NewSeparator(),
		widget.NewButton("Back", tv.onClose),
	)

	txt := widgets.NewBidiText()
	txt.SetText(project.Target.Paragraphs[0].Text)

	content := container.NewBorder(
		nil,
		toolBar,
		nil,
		nil,
		txt,
	)

	tv.View = content

	return tv
}

func (tv *TranslationView) onClose() {
	if tv.OnClose != nil {
		tv.OnClose()
	}
}

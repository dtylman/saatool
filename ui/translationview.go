package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

type TranslationView struct {
	OnClose func()
	View    fyne.CanvasObject
	project *translation.Project
	txt     *widgets.BidiText
}

func NewTranslationView(project *translation.Project) *TranslationView {
	tv := &TranslationView{
		project: project,
	}
	backBtn := widget.NewButton("Close", tv.onClose)

	txt := widgets.NewBidiText(project.Target.Paragraphs[0].Text)

	content := container.NewBorder(
		nil,
		backBtn,
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

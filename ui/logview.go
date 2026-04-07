package ui

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/ui/widgets"
)

// LogView represents the log view in the application.
type LogView struct {
	view     fyne.CanvasObject
	list     *widget.List
	messages []string
}

// NewLogView creates a new LogView instance.
func NewLogView() *LogView {
	lv := &LogView{}

	lv.list = widget.NewList(
		func() int {
			return len(lv.messages)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapBreak
			label.TextStyle = fyne.TextStyle{Monospace: true}
			return label
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			label.SetText(lv.messages[id])
			height := label.MinSize().Height
			lv.list.SetItemHeight(id, height)
		},
	)

	lv.view = lv.list

	Main.ClearActions()
	Main.AddAction("Refresh", widgets.IconReload, lv.refreshTapped)

	return lv
}

func (lv *LogView) View() fyne.CanvasObject {
	return lv.view
}

func (lv *LogView) Close() {
	// nothing to do
}

func (lv *LogView) Load() {
	lv.loadMessages()
}

func (lv *LogView) loadMessages() {
	lv.messages = MemoryLog.GetMessages()
	slices.Reverse(lv.messages)
	lv.list.Refresh()
}

func (lv *LogView) refreshTapped() {
	lv.loadMessages()
}

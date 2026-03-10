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
	textGrid *widget.TextGrid
	messages []string
}

// NewLogView creates a new LogView instance.
func NewLogView() *LogView {
	lv := &LogView{}

	lv.textGrid = widget.NewTextGrid()
	lv.textGrid.ShowLineNumbers = true
	lv.textGrid.Scroll = fyne.ScrollBoth
	lv.textGrid.ShowWhitespace = false

	lv.view = lv.textGrid

	return lv
}

func (lv *LogView) View() fyne.CanvasObject {
	return lv.view
}

func (lv *LogView) Close() {}

// Load is called when the Log tab becomes active.
func (lv *LogView) Load() {
	Main.ClearActions()
	Main.AddAction("Refresh", widgets.IconReload, lv.refreshTapped)
	lv.loadMessages()
	lv.textGrid.Refresh()
}

func (lv *LogView) loadMessages() {
	lv.messages = MemoryLog.GetMessages()
	slices.Reverse(lv.messages)
	lv.textGrid.Rows = make([]widget.TextGridRow, 0)
	for _, msg := range lv.messages {
		lv.textGrid.Append(msg)
	}
	lv.textGrid.Refresh()
}

func (lv *LogView) refreshTapped() {
	lv.loadMessages()
	lv.textGrid.Refresh()
}

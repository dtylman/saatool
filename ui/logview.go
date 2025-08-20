package ui

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/ui/widgets"
)

// LogView represents the log view in the application.
type LogView struct {
	View     fyne.CanvasObject
	table    *widget.Table
	messages []string
}

// NewLogView creates a new LogView instance.
func NewLogView() *LogView {
	lv := &LogView{}

	lv.table = widget.NewTable(
		lv.len, lv.create, lv.update,
	)

	lv.View = lv.table

	lv.loadMessages()

	Main.ClearActions()
	Main.AddAction("Refresh", widgets.IconReload, lv.refreshTapped)

	return lv
}

func (lv *LogView) len() (int, int) {
	return len(lv.messages), 1
}

func (lv *LogView) create() fyne.CanvasObject {
	return widget.NewLabel("")
}

func (lv *LogView) update(id widget.TableCellID, cell fyne.CanvasObject) {
	if id.Row < 0 || id.Row >= len(lv.messages) {
		return
	}

	label := cell.(*widget.Label)
	label.SetText(lv.messages[id.Row])
}

func (lv *LogView) loadMessages() {
	lv.messages = MemoryLog.GetMessages()
	slices.Reverse(lv.messages)
}

func (lv *LogView) refreshTapped() {
	lv.loadMessages()
	lv.table.Refresh()
}

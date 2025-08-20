package ui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LogView struct {
	View     fyne.CanvasObject
	lstLog   *widget.List
	mutex    sync.Mutex
	messages []string
	OnLog    func(string)
}

func NewLogView() *LogView {
	lv := &LogView{
		messages: make([]string, 0),
	}

	lv.lstLog = widget.NewList(lv.len, lv.createItem, lv.updateItem)
	lv.View = container.NewVScroll(lv.lstLog)
	return lv
}

func (lv *LogView) Write(p []byte) (n int, err error) {
	lv.addMessage(string(p))
	// fyne.Do(func() {
	// 	if lv.lstLog == nil {
	// 		return
	// 	}
	// 	lv.lstLog.Refresh()
	// })
	return len(p), nil
}

func (lv *LogView) addMessage(msg string) {
	lv.mutex.Lock()
	defer lv.mutex.Unlock()

	lv.messages = append(lv.messages, msg)
	if len(lv.messages) > 1000 { // Limit to last 1000 messages
		lv.messages = lv.messages[len(lv.messages)-1000:]
	}
}
func (lv *LogView) len() int {
	lv.mutex.Lock()
	defer lv.mutex.Unlock()
	return len(lv.messages)
}

func (lv *LogView) createItem() fyne.CanvasObject {
	lv.mutex.Lock()
	defer lv.mutex.Unlock()

	label := widget.NewLabel("")
	label.Wrapping = fyne.TextWrapWord
	label.Selectable = false
	label.TextStyle.Monospace = true

	return label
}

func (lv *LogView) updateItem(id widget.ListItemID, item fyne.CanvasObject) {
	lv.mutex.Lock()
	defer lv.mutex.Unlock()

	if id < 0 || id >= len(lv.messages) {
		return
	}

	label := item.(*widget.Label)
	label.SetText(lv.messages[id])
}

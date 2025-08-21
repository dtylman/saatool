package ui

import (
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/ui/widgets"
)

type SettingsView struct {
	preferences *PreferencesDecorator

	entryDeepSeekAPIKey *widget.Entry
	entryTextSize       *widget.Entry

	View fyne.CanvasObject
}

func NewSettingsView(preferences *PreferencesDecorator) *SettingsView {
	sv := &SettingsView{
		entryDeepSeekAPIKey: widget.NewEntry(),
		entryTextSize:       widget.NewEntry(),
		preferences:         preferences,
	}

	sv.View = widget.NewForm(
		widget.NewFormItem("DeepSeek API Key", sv.entryDeepSeekAPIKey),
		widget.NewFormItem("App Sizes Factor", sv.entryTextSize),
	)

	Main.ClearActions()
	Main.AddAction("Save", widgets.IconSave, sv.onSaveTapped)

	sv.entryDeepSeekAPIKey.SetText(preferences.DeepSeekAPIKey())
	sv.entryTextSize.SetText(fmt.Sprintf("%v", preferences.AppSize()))

	return sv

}

func (sv *SettingsView) onSaveTapped() {
	sv.preferences.SetDeepSeekAPIKey(sv.entryDeepSeekAPIKey.Text)
	newSize, err := strconv.Atoi(sv.entryTextSize.Text)
	if err != nil {
		log.Printf("invalid app size: %v", err)
		sv.entryTextSize.SetText(fmt.Sprintf("%v", sv.preferences.AppSize()))
	} else {
		sv.preferences.SetAppSize(newSize)
	}

}

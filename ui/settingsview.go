package ui

import (
	"fmt"

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
		widget.NewFormItem("Translation Text Size", sv.entryTextSize),
	)

	Main.ClearActions()
	Main.AddAction("Save", widgets.IconSave, sv.onSaveTapped)

	sv.entryDeepSeekAPIKey.SetText(preferences.DeepSeekAPIKey())
	sv.entryTextSize.SetText(fmt.Sprintf("%.2f", preferences.TranslationTextSize()))

	return sv

}

func (sv *SettingsView) onSaveTapped() {
	sv.preferences.SetDeepSeekAPIKey(sv.entryDeepSeekAPIKey.Text)
}

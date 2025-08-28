package ui

import (
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/ui/widgets"
)

type SettingsView struct {
	entryDeepSeekAPIKey *widget.Entry
	entryTextSize       *widget.Entry

	view fyne.CanvasObject
}

func NewSettingsView() *SettingsView {
	sv := &SettingsView{
		entryDeepSeekAPIKey: widget.NewEntry(),
		entryTextSize:       widget.NewEntry(),
	}

	sv.view = widget.NewForm(
		widget.NewFormItem("DeepSeek API Key", sv.entryDeepSeekAPIKey),
		widget.NewFormItem("App Sizes Factor", sv.entryTextSize),
	)

	Main.ClearActions()
	Main.AddAction("Save", widgets.IconSave, sv.onSaveTapped)

	sv.entryDeepSeekAPIKey.SetText(config.Options.DeepSeekAPIKey)
	sv.entryTextSize.SetText(fmt.Sprintf("%v", config.Options.AppSize))

	return sv

}

func (sv *SettingsView) View() fyne.CanvasObject {
	return sv.view
}

func (sv *SettingsView) Close() {
	// nothing to do
}

func (sv *SettingsView) Load() {
	// nothing to do
}

func (sv *SettingsView) onSaveTapped() {
	config.Options.DeepSeekAPIKey = sv.entryDeepSeekAPIKey.Text
	newSize, err := strconv.Atoi(sv.entryTextSize.Text)
	if err != nil {
		log.Printf("invalid app size: %v", err)
		sv.entryTextSize.SetText(fmt.Sprintf("%v", config.Options.AppSize))
	} else {
		config.Options.AppSize = newSize
	}

	err = config.SaveOptions()
	if err != nil {
		log.Printf("failed to save settings: %v", err)
		Main.ShowError(fmt.Sprintf("failed to save settings: %v", err))
	}
}

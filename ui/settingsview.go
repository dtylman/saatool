package ui

import (
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/ui/widgets"
)

type SettingsView struct {
	entryDeepSeekAPIKey *widget.Entry
	entryAppSize        *widget.Entry
	entryTransDocSize   *widget.Entry
	entryTranslateAhead *widget.Entry
	entryAutoProofread  *widget.Check
	entrySourceLanguage *widget.Entry
	entryTargetLanguage *widget.Entry
	entryDarkMode       *widget.Check

	view fyne.CanvasObject
}

func NewSettingsView() *SettingsView {
	sv := &SettingsView{
		entryDeepSeekAPIKey: widget.NewEntry(),
		entryAppSize:        widget.NewEntry(),
		entryTranslateAhead: widget.NewEntry(),
		entryTransDocSize:   widget.NewEntry(),
		entryAutoProofread:  widget.NewCheck("", nil),
		entrySourceLanguage: widget.NewEntry(),
		entryTargetLanguage: widget.NewEntry(),
		entryDarkMode:       widget.NewCheck("", nil),
	}

	sv.entryDeepSeekAPIKey.SetText(config.Options.DeepSeekAPIKey)
	sv.entryDeepSeekAPIKey.Password = true
	sv.entryTranslateAhead.SetText(fmt.Sprintf("%v", config.Options.TranslateAhead))
	sv.entryAutoProofread.SetChecked(config.Options.AutoProofread)
	sv.entryAppSize.SetText(fmt.Sprintf("%v", config.Options.AppSize))
	sv.entryTransDocSize.SetText(fmt.Sprintf("%v", config.Options.TranslationDocSize))
	sv.entrySourceLanguage.SetText(config.Options.SourceLanguage)
	sv.entryTargetLanguage.SetText(config.Options.TargetLanguage)
	sv.entryDarkMode.SetChecked(config.Options.DarkMode)

	sv.view = container.NewPadded(container.NewVBox(
		buildSettingsSection("Appearance",
			widget.NewFormItem("Dark Mode", sv.entryDarkMode),
		),
		buildSettingsSection("Languages",
			widget.NewFormItem("Source Language", sv.entrySourceLanguage),
			widget.NewFormItem("Target Language", sv.entryTargetLanguage),
		),
		buildSettingsSection("AI / API",
			widget.NewFormItem("DeepSeek API Key", sv.entryDeepSeekAPIKey),
		),
		buildSettingsSection("Translation",
			widget.NewFormItem("Translate Ahead", sv.entryTranslateAhead),
			widget.NewFormItem("Auto Proofread", sv.entryAutoProofread),
			widget.NewFormItem("Doc Size", sv.entryTransDocSize),
			widget.NewFormItem("App Size Factor", sv.entryAppSize),
		),
	))

	return sv
}

// buildSettingsSection creates a labeled group: bold header + separator + form items.
func buildSettingsSection(title string, items ...*widget.FormItem) fyne.CanvasObject {
	header := canvas.NewText(title, theme.Color(theme.ColorNameForeground))
	header.TextStyle = fyne.TextStyle{Bold: true}
	header.TextSize = theme.Size(theme.SizeNameText) + 2

	sep := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	sep.SetMinSize(fyne.NewSize(0, 1))

	form := widget.NewForm(items...)

	return container.NewVBox(
		container.NewPadded(header),
		sep,
		form,
	)
}

func (sv *SettingsView) View() fyne.CanvasObject {
	return sv.view
}

func (sv *SettingsView) Close() {}

func (sv *SettingsView) Load() {
	Main.ClearActions()
	Main.AddAction("Save", widgets.IconSave, sv.onSaveTapped)
}

func (sv *SettingsView) onSaveTapped() {
	config.Options.DeepSeekAPIKey = sv.entryDeepSeekAPIKey.Text

	newSize, err := strconv.Atoi(sv.entryAppSize.Text)
	if err != nil {
		log.Printf("invalid app size: %v", err)
		sv.entryAppSize.SetText(fmt.Sprintf("%v", config.Options.AppSize))
	} else {
		config.Options.AppSize = newSize
	}

	newTranslateAhead, err := strconv.Atoi(sv.entryTranslateAhead.Text)
	if err != nil {
		log.Printf("invalid translate ahead: %v", err)
		sv.entryTranslateAhead.SetText(fmt.Sprintf("%v", config.Options.TranslateAhead))
	} else {
		config.Options.TranslateAhead = newTranslateAhead
	}

	config.Options.AutoProofread = sv.entryAutoProofread.Checked

	newDocSize, err := strconv.Atoi(sv.entryTransDocSize.Text)
	if err != nil {
		log.Printf("invalid translation document size: %v", err)
		sv.entryTransDocSize.SetText(fmt.Sprintf("%v", config.Options.TranslationDocSize))
	} else {
		config.Options.TranslationDocSize = newDocSize
	}

	config.Options.SourceLanguage = sv.entrySourceLanguage.Text
	config.Options.TargetLanguage = sv.entryTargetLanguage.Text
	config.Options.DarkMode = sv.entryDarkMode.Checked

	err = config.SaveOptions()
	if err != nil {
		log.Printf("failed to save settings: %v", err)
		Main.ShowError(fmt.Sprintf("failed to save settings: %v", err))
	} else {
		Main.ApplyTheme()
		Main.ShowMessage("Settings saved")
	}
}

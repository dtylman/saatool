package ui

import (
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/ui/widgets"
)

type SettingsView struct {
	selectAIVendor      *widget.Select
	selectAIModel       *widget.Select
	entryAIAPIKey       *widget.Entry
	entryAppSize        *widget.Entry
	entryTransDocSize   *widget.Entry
	entryTranslateAhead *widget.Entry
	entryAutoProofread  *widget.Check

	view fyne.CanvasObject
}

func NewSettingsView() *SettingsView {
	vendors := ai.SupportedVendors
	if len(vendors) == 0 {
		vendors = []string{"deepseek"}
	}

	sv := &SettingsView{
		selectAIVendor:      widget.NewSelect(vendors, nil),
		selectAIModel:       widget.NewSelect([]string{}, nil),
		entryAIAPIKey:       widget.NewEntry(),
		entryAppSize:        widget.NewEntry(),
		entryTranslateAhead: widget.NewEntry(),
		entryTransDocSize:   widget.NewEntry(),
		entryAutoProofread:  widget.NewCheck("", nil),
	}

	sv.selectAIVendor.OnChanged = func(vendor string) {
		sv.setModelsForVendor(vendor)
	}

	sv.view = widget.NewForm(
		widget.NewFormItem("AI Vendor", sv.selectAIVendor),
		widget.NewFormItem("AI Model", sv.selectAIModel),
		widget.NewFormItem("AI API Key", sv.entryAIAPIKey),
		widget.NewFormItem("App Sizes Factor", sv.entryAppSize),
		widget.NewFormItem("Translate Ahead", sv.entryTranslateAhead),
		widget.NewFormItem("Auto Proofread", sv.entryAutoProofread),
		widget.NewFormItem("Translation Doc Size", sv.entryTransDocSize),
	)

	Main.ClearActions()
	Main.AddAction("Save", widgets.IconSave, sv.onSaveTapped)

	vendor := config.Options.AI.Vendor
	if vendor == "" {
		vendor = vendors[0]
	}
	sv.selectAIVendor.SetSelected(vendor)

	sv.setModelsForVendor(vendor)
	model := config.Options.AI.Model
	if model == "" && len(sv.selectAIModel.Options) > 0 {
		model = sv.selectAIModel.Options[0]
	}
	sv.selectAIModel.SetSelected(model)

	sv.entryAIAPIKey.SetText(config.Options.AI.APIKey)
	sv.entryAIAPIKey.Password = true
	sv.entryTranslateAhead.SetText(fmt.Sprintf("%v", config.Options.TranslateAhead))
	sv.entryAutoProofread.SetChecked(config.Options.AutoProofread)

	sv.entryAppSize.SetText(fmt.Sprintf("%v", config.Options.AppSize))
	sv.entryTransDocSize.SetText(fmt.Sprintf("%v", config.Options.TranslationDocSize))

	return sv

}

func (sv *SettingsView) setModelsForVendor(vendor string) {
	models := ai.SupportedModels(vendor)
	if len(models) == 0 {
		models = []string{""}
	}

	previousModel := sv.selectAIModel.Selected
	sv.selectAIModel.Options = models
	if previousModel != "" {
		for _, model := range models {
			if model == previousModel {
				sv.selectAIModel.SetSelected(previousModel)
				return
			}
		}
	}
	sv.selectAIModel.SetSelected(models[0])
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
	config.Options.AI.Vendor = sv.selectAIVendor.Selected
	config.Options.AI.Model = sv.selectAIModel.Selected
	config.Options.AI.APIKey = sv.entryAIAPIKey.Text
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
	err = config.SaveOptions()
	if err != nil {
		log.Printf("failed to save settings: %v", err)
		Main.ShowError(fmt.Sprintf("failed to save settings: %v", err))
	} else {
		Main.ShowMessage("Settings saved")
	}
}

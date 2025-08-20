package ui

import "fyne.io/fyne/v2"

type PreferencesDecorator struct {
	prefs fyne.Preferences
}

func NewPreferencesDecorator(prefs fyne.Preferences) *PreferencesDecorator {
	return &PreferencesDecorator{prefs: prefs}
}

func (pd *PreferencesDecorator) SetActiveProject(projectPath string) {
	pd.prefs.SetString("active_project", projectPath)
}

func (pd *PreferencesDecorator) ActiveProject() string {
	return pd.prefs.String("active_project")
}

func (pd *PreferencesDecorator) TranslationTextSize() float32 {
	return float32(pd.prefs.FloatWithFallback("translation_text_size", 20))
}

func (pd *PreferencesDecorator) TranslationTextPadding() float32 {
	return float32(pd.prefs.FloatWithFallback("translation_text_padding", 5))
}

func (pd *PreferencesDecorator) TranslationTextSpacing() float32 {
	return float32(pd.prefs.FloatWithFallback("translation_text_spacing", 5))
}

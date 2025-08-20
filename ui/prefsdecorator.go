package ui

import "fyne.io/fyne/v2"

type PreferencesDecorator struct {
	prefs fyne.Preferences
}

func NewPreferencesDecorator(prefs fyne.Preferences) *PreferencesDecorator {
	return &PreferencesDecorator{prefs: prefs}
}

// DeepSeekAPIKey retrieves the DeepSeek API key from preferences.
func (pd *PreferencesDecorator) DeepSeekAPIKey() string {
	return pd.prefs.StringWithFallback("deepseek_api_key", "")
}

// SetDeepSeekAPIKey sets the DeepSeek API key in preferences.
func (pd *PreferencesDecorator) SetDeepSeekAPIKey(apiKey string) {
	pd.prefs.SetString("deepseek_api_key", apiKey)
}

// ActiveProject sets the active project path in preferences.
func (pd *PreferencesDecorator) SetActiveProject(projectPath string) {
	pd.prefs.SetString("active_project", projectPath)
}

// ActiveProject retrieves the active project path from preferences.
func (pd *PreferencesDecorator) ActiveProject() string {
	return pd.prefs.StringWithFallback("active_project", "")
}

// TranslationTextSize retrieves the translation text size from preferences.
func (pd *PreferencesDecorator) TranslationTextSize() float32 {
	return float32(pd.prefs.FloatWithFallback("translation_text_size", 20))
}

// TranslationTextPadding retrieves the translation text padding from preferences.
func (pd *PreferencesDecorator) TranslationTextPadding() float32 {
	return float32(pd.prefs.FloatWithFallback("translation_text_padding", 5))
}

// TranslationTextSpacing retrieves the translation text spacing from preferences.
func (pd *PreferencesDecorator) TranslationTextSpacing() float32 {
	return float32(pd.prefs.FloatWithFallback("translation_text_spacing", 5))
}

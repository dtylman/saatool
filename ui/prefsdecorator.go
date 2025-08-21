package ui

import "fyne.io/fyne/v2"

type PreferencesDecorator struct {
	prefs fyne.Preferences
}

func (pd *PreferencesDecorator) AppSize() int {
	return pd.prefs.IntWithFallback("app_size", 16)
}

func (pd *PreferencesDecorator) SetAppSize(size int) {
	pd.prefs.SetInt("app_size", size)
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

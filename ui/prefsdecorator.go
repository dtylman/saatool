package ui

import "fyne.io/fyne/v2"

// PreferencesDecorator is a decorator for fyne.Preferences that provides additional methods
type PreferencesDecorator struct {
	prefs fyne.Preferences
}

// LastSource retrieves the last source language preference.
func (pd *PreferencesDecorator) LastTranslationSource() bool {
	return pd.prefs.BoolWithFallback("last_translation_displayed_source", false)
}
func (pd *PreferencesDecorator) SetLastTranslationSource(source bool) {
	pd.prefs.SetBool("last_translation_displayed_source", source)
}

// LastParagraph retrieves the last displayed paragraph index.
func (pd *PreferencesDecorator) LastTranslationParagraph() int {
	return pd.prefs.IntWithFallback("last_translation_displayed_paragraph", 0)
}

// SetLastParagraph sets the last displayed paragraph index.
func (pd *PreferencesDecorator) SetLastTranslationParagraph(paragraph int) {
	pd.prefs.SetInt("last_translation_displayed_paragraph", paragraph)
}
func NewPreferencesDecorator(prefs fyne.Preferences) *PreferencesDecorator {
	return &PreferencesDecorator{prefs: prefs}
}

// ActiveProject sets the active project path in preferences.
func (pd *PreferencesDecorator) SetActiveProject(projectPath string) {
	pd.prefs.SetString("active_project", projectPath)
}

// ActiveProject retrieves the active project path from preferences.
func (pd *PreferencesDecorator) ActiveProject() string {
	return pd.prefs.StringWithFallback("active_project", "")
}

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
)

// MainWindow represents the main window of the SaaTool application.
type MainWindow struct {
	app    fyne.App
	Window fyne.Window
	tabs   *container.AppTabs
}

// NewMainWindow creates a new instance of the main window
func NewMainWindow(a fyne.App) *MainWindow {
	w := a.NewWindow("SaaTool Main Window")
	w.Resize(fyne.NewSize(800, 600))
	w.SetMaster()

	// Placeholder views for each module
	mw := &MainWindow{
		Window: w,
		app:    a,
	}

	projectView := NewProjectsView()
	projectView.OnProjectSelected = mw.ShowTranslationView

	settingsView := widget.NewLabel("Settings Page (to be implemented)")

	// Only Projects and Settings tabs by default
	tabs := container.NewAppTabs(
		container.NewTabItem("Projects", projectView.View),
		container.NewTabItem("Settings", settingsView),
	)
	tabs.SetTabLocation(container.TabLocationBottom)

	w.SetContent(tabs)
	mw.tabs = tabs

	return mw
}

// ShowTranslationView switches the window to the translation view for the selected project
func (mw *MainWindow) ShowTranslationView(project *translation.Project) {
	tv := NewTranslationView(project)
	tv.OnClose = func() {
		mw.Window.SetContent(mw.tabs)
	}
	mw.Window.SetContent(tv.View)
}

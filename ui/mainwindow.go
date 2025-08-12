package ui

import "fyne.io/fyne/v2"

// MainWindow represents the main window of the SaaTool application.
type MainWindow struct {
	app    fyne.App
	Window fyne.Window
}

// NewMainWindow creates a new instance of the main window
func NewMainWindow(a fyne.App) *MainWindow {
	w := a.NewWindow("SaaTool Main Window")
	w.Resize(fyne.NewSize(800, 600))

	w.SetMaster()

	return &MainWindow{
		Window: w,
		app:    a,
	}
}

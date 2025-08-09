package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MainWindow represents the main window of the SaaTool application.
type MainWindow struct {
	App    fyne.App
	Window fyne.Window
	Label  *widget.Label
}

// NewMainWindow creates a new instance of the main window
func NewMainWindow() *MainWindow {
	a := app.New()
	w := a.NewWindow("SaaTool Main Window")
	w.Resize(fyne.NewSize(800, 600))

	label := widget.NewLabel("Welcome to SaaTool!")
	content := container.NewVBox(
		label,
		widget.NewButton("Click Me", func() {
			label.SetText("Button clicked!")
		}),
	)

	w.SetContent(content)

	return &MainWindow{
		App:    a,
		Window: w,
		Label:  label,
	}
}

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
	// w.Resize(fyne.NewSize(200, 300))

	label := widget.NewLabel("Welcome to SaaTool!")

	hebrewText := `יום יום אני תולש מהלוח דף. יום ראשון - כמעט. יום שני - I'm happy! ויום שלישי - 365 ימים בשנה!`
	hebrewText += "\n" // Adding a newline for better visibility
	hebrewText += "This is a test of RTL text rendering."
	hebrewText += "\n" // Adding another newline for clarity
	hebrewText += "שלום שלום נתראה בחלום. אני יושב על הכיסא ומחכה לך."
	rtlWidget := NewBidiLabel(hebrewText)

	content := container.NewVBox(
		label,
		rtlWidget,
		NewBidiText(sampleText),
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

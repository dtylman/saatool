package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("saatool")
	w := a.NewWindow("saatool")
	w.Resize(fyne.NewSize(640, 480))
	w.CenterOnScreen()

	closeBtn := widget.NewButton("Close", func() {
		//dialog.ShowInformation("Info", "This is a simple Fyne application.", w)
		w.Close()
	})

	bidiText := NewBidiText(sampleText)
	w.SetContent(
		container.NewBorder(
			nil, // top
			container.NewHBox(
				layout.NewSpacer(),
				closeBtn,
			), // bottom
			nil, // left
			nil, // right
			bidiText,
		),
	)

	w.ShowAndRun()
}

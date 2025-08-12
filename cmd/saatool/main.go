package main

import (
	"log"

	"fyne.io/fyne/v2/app"
	"github.com/dtylman/saatool/ui"
	"github.com/dtylman/saatool/ui/widgets"
)

func main() {
	log.Printf("Starting SaaTool application...")
	a := app.New()
	a.Settings().SetTheme(&widgets.Theme{})
	window := ui.NewMainWindow(a)
	window.Window.ShowAndRun()
}

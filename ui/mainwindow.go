package ui

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/ui/widgets"
)

// Main is the global instance of the main application window.
var Main *MainWindow

// MainWindow represents the main application window.
type MainWindow struct {
	fyneApp fyne.App
	window  fyne.Window
	toolBar *fyne.Container
}

func NewMainWindow() error {
	if Main != nil {
		return errors.New("main window already exists")
	}
	Main = &MainWindow{
		fyneApp: app.New(),
		window:  nil,
		toolBar: container.NewHBox(),
	}
	return nil
}

// ShowAndRun creates the main application window and starts the Fyne event loop.
func (mw *MainWindow) ShowAndRun() {

	mw.fyneApp = app.New()
	mw.fyneApp.Settings().SetTheme(&widgets.Theme{})

	mw.window = mw.fyneApp.NewWindow("SaaTool")

	mw.window.Resize(fyne.NewSize(800, 600))
	mw.window.SetMaster()

	mw.onProjectsTapped()

	mw.window.ShowAndRun()
}

// SetContent sets the content of the main window.
func (mw *MainWindow) SetContent(content fyne.CanvasObject) {
	panelTop := container.NewHBox(
		widget.NewIcon(widgets.LoadIcon),
		widget.NewLabel("SaaTool"),
	)

	btnTranslate := widget.NewButtonWithIcon("Projects", widgets.LoadIcon, mw.onProjectsTapped)
	btnSettings := widget.NewButtonWithIcon("Settings", widgets.LoadIcon, mw.onSettingsTapped)

	panelBottom := container.NewVBox(
		mw.toolBar,
		widgets.NewPanel(
			container.NewHBox(
				btnTranslate,
				btnSettings,
			), fyne.NewSize(0, 50),
		),
		container.NewHBox(
			widget.NewLabel("Status: Ready"),
		),
	)

	mw.window.SetContent(
		container.NewBorder(
			panelTop,
			panelBottom,
			nil,
			nil,
			container.NewVScroll(content),
		),
	)

	content.Show()
	content.Refresh()
}

func (mw *MainWindow) ClearActions() {
	mw.toolBar.RemoveAll()
	mw.toolBar.Refresh()
}

func (mw *MainWindow) AddActionWidget(widget fyne.CanvasObject) {
	mw.toolBar.Add(widget)
	mw.toolBar.Refresh()
}

func (mw *MainWindow) AddAction(label string, icon fyne.Resource, action func()) {
	btn := widget.NewButtonWithIcon(label, icon, action)
	mw.toolBar.Add(btn)
}

func (mw *MainWindow) onSettingsTapped() {
	mw.SetContent(
		widget.NewLabel("Settings Page (to be implemented)"),
	)
}

func (mw *MainWindow) onProjectsTapped() {
	mw.SetContent(
		NewProjectsView().View,
	)
}

func (mw *MainWindow) ShowMessage(message string) {
	dialog := widget.NewPopUp(
		widget.NewLabel(message),
		mw.window.Canvas(),
	)
	dialog.Show()
	dialog.Resize(fyne.NewSize(300, 100))
	dialog.Move(fyne.NewPos(
		mw.window.Canvas().Size().Width/2-150,
		mw.window.Canvas().Size().Height/2-50,
	))
}

func (mw *MainWindow) ShowError(message string) {
	dialog := widget.NewPopUp(
		widget.NewLabel("Error: "+message),
		mw.window.Canvas(),
	)
	dialog.Show()
	dialog.Resize(fyne.NewSize(300, 100))
	dialog.Move(fyne.NewPos(
		mw.window.Canvas().Size().Width/2-150,
		mw.window.Canvas().Size().Height/2-50,
	))
}

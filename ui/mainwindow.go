package ui

import (
	"errors"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"fyne.io/x/fyne/layout"
	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/ui/widgets"
)

// Main is the global instance of the main application window.
var Main *MainWindow

// WindowContent is an interface for views that can be shown in the main area
type WindowContent interface {
	View() fyne.CanvasObject
	Close()
	Load()
}

// MainWindow represents the main application window.
type MainWindow struct {
	fyneApp    fyne.App
	window     fyne.Window
	content    WindowContent
	toolBar    *fyne.Container
	header     *widget.Label
	translator *ai.Translator
}

func (mw *MainWindow) OpenSaveDialog(callback func(fyne.URIWriteCloser, error), filter ...string) {
	fg := dialog.NewFileSave(callback, mw.window)
	fg.SetFilter(storage.NewExtensionFileFilter(filter))
	fg.Show()
}

// OpenFileDialog opens a file dialog to select a file and calls the callback with the selected file.
func (mw *MainWindow) OpenFileDialog(callback func(reader fyne.URIReadCloser, err error), filter ...string) {
	fd := dialog.NewFileOpen(callback, mw.window)
	fd.SetFilter(storage.NewExtensionFileFilter(filter))
	fd.Show()
}

// NewMAinWindow creates a new instance of the main application window.
func NewMainWindow() error {
	if Main != nil {
		return errors.New("main window already exists")
	}

	Main = &MainWindow{
		fyneApp: app.New(),
		window:  nil,
		toolBar: container.NewGridWrap(fyne.NewSize(100, 40)),
		header:  widget.NewLabel(fmt.Sprintf("SaaTool %v-%v", app.New().Metadata().Version, app.New().Metadata().Build)),
	}

	return nil
}

// ShowAndRun creates the main application window and starts the Fyne event loop.
func (mw *MainWindow) ShowAndRun() {
	defer log.Println("exiting application")
	mw.fyneApp = app.New()

	err := config.LoadOptions()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	mw.fyneApp.Settings().SetTheme(widgets.NewTheme(config.Options.AppSize))

	mw.window = mw.fyneApp.NewWindow("SaaTool")

	mw.window.Resize(fyne.NewSize(800, 600))
	mw.window.SetMaster()

	mw.onProjectsTapped()

	mw.window.ShowAndRun()

}

// SetContent sets the content of the main window.
func (mw *MainWindow) SetContent(content WindowContent) {
	if mw.content != nil {
		mw.content.Close()
	}
	mw.content = content
	mw.Refresh()
}

func (mw *MainWindow) Refresh() {
	if mw.content == nil {
		return
	}
	fyne.Do(func() {
		panelTop := container.NewHBox(
			widget.NewIcon(widgets.IconLogo),
			mw.header,
		)

		panelBottom := container.NewVBox(
			mw.toolBar,

			layout.NewResponsiveLayout(
				layout.Responsive(widget.NewButtonWithIcon("Projects", widgets.IconProject, mw.onProjectsTapped), 0.33),
				layout.Responsive(widget.NewButtonWithIcon("Settings", widgets.IconSettings, mw.onSettingsTapped), 0.33),
				layout.Responsive(widget.NewButtonWithIcon("Log", widgets.IconLog, mw.onLogTapped), 0.33),
			),
		)

		mw.window.SetContent(
			container.NewBorder(
				panelTop,
				panelBottom,
				nil,
				nil,
				container.NewVScroll(mw.content.View()),
			),
		)

		mw.content.Load()
	})
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
	mw.SetContent(NewSettingsView())
}

func (mw *MainWindow) onProjectsTapped() {
	mw.SetContent(NewProjectsView())
}

func (mw *MainWindow) onLogTapped() {
	mw.SetContent(NewLogView())
}

func (mw *MainWindow) ShowMessage(message string) {
	fyne.Do(dialog.NewInformation("Message", message, mw.window).Show)
}

func (mw *MainWindow) ShowError(message string) {
	fyne.Do(dialog.NewError(errors.New(message), mw.window).Show)
}

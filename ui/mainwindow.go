package ui

import (
	"errors"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"fyne.io/x/fyne/layout"
	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/ui/widgets"
)

// Main is the global instance of the main application window.
var Main *MainWindow

// MainWindow represents the main application window.
type MainWindow struct {
	fyneApp    fyne.App
	window     fyne.Window
	toolBar    *fyne.Container
	translator *ai.Translator
	logView    *LogView
	txtStatus  *widget.Label
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
		fyneApp:   app.New(),
		window:    nil,
		toolBar:   container.NewHBox(),
		logView:   NewLogView(),
		txtStatus: widget.NewLabel("Status: Ready"),
	}

	Main.txtStatus.Truncation = fyne.TextTruncateClip
	Main.txtStatus.Wrapping = fyne.TextWrapOff

	// bind the log view to the main window
	// teeWriter := io.MultiWriter(Main.logView, log.Writer())
	//log.SetOutput(teeWriter)
	log.SetOutput(Main.logView)
	// Main.logView.OnLog = Main.onLogMessage

	return nil
}

// ShowAndRun creates the main application window and starts the Fyne event loop.
func (mw *MainWindow) ShowAndRun() {

	mw.fyneApp = app.New()
	mw.fyneApp.Settings().SetTheme(&widgets.Theme{})

	mw.window = mw.fyneApp.NewWindow("SaaTool")

	mw.window.Resize(fyne.NewSize(800, 600))
	mw.window.SetMaster()

	mw.onProjectTapped()

	mw.window.ShowAndRun()
}

// SetContent sets the content of the main window.
func (mw *MainWindow) SetContent(content fyne.CanvasObject) {
	fyne.Do(func() {
		panelTop := container.NewHBox(
			widget.NewIcon(widgets.IconLogo),
			widget.NewLabel("SaaTool"),
		)

		panelBottom := container.NewVBox(
			mw.toolBar,

			layout.NewResponsiveLayout(
				layout.Responsive(widget.NewButtonWithIcon("Project", widgets.IconProject, mw.onProjectTapped), 0.33),
				layout.Responsive(widget.NewButtonWithIcon("Settings", widgets.IconSettings, mw.onSettingsTapped), 0.33),
				layout.Responsive(widget.NewButtonWithIcon("Log", widgets.IconLog, mw.onLogTapped), 0.33),
			),

			container.NewHBox(
				mw.txtStatus,
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
	mw.SetContent(
		widget.NewLabel("Settings Page (to be implemented)"),
	)
}

func (mw *MainWindow) onProjectTapped() {
	mw.SetContent(
		NewProjectView().View,
	)
}

func (mw *MainWindow) onLogTapped() {
	mw.ClearActions()
	mw.SetContent(
		mw.logView.View,
	)
}

func (mw *MainWindow) ShowMessage(message string) {
	fyne.Do(dialog.NewInformation("Message", message, mw.window).Show)
}

func (mw *MainWindow) ShowError(message string) {
	fyne.Do(dialog.NewError(errors.New(message), mw.window).Show)
}

func (mw *MainWindow) Preferences() *PreferencesDecorator {
	return NewPreferencesDecorator(mw.fyneApp.Preferences())
}

func (mw *MainWindow) Translator() *ai.Translator {
	if mw.translator == nil {
		mw.translator = ai.NewTranslator()
	}
	return mw.translator
}

func (mw *MainWindow) onLogMessage(msg string) {
	fyne.Do(func() {
		mw.txtStatus.SetText(msg)
	})
}

package ui

import (
	"errors"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// Main is the global instance of the main application window.
var Main *MainWindow

// WindowContent is an interface for views that can be shown in the main area.
type WindowContent interface {
	View() fyne.CanvasObject
	Close()
	Load()
}

// MainWindow represents the main application window.
type MainWindow struct {
	fyneApp fyne.App
	window  fyne.Window
	toolBar *fyne.Container // per-view action buttons (populated by each view's Load)
	reader  WindowContent   // non-nil when in fullscreen reading mode
}

func (mw *MainWindow) OpenProjectSaveDialog(callback func(fyne.URIWriteCloser, error), project *translation.Project) {
	fg := dialog.NewFileSave(callback, mw.window)
	if project != nil {
		fileName := project.Name
		if fileName == "" {
			fileName = "untitled"
		}
		if !strings.HasSuffix(fileName, config.ProjectFileExt) {
			fileName += config.ProjectFileExt
		}
		fg.SetFileName(fileName)
	}
	fg.SetFilter(storage.NewExtensionFileFilter([]string{config.ProjectFileExt}))
	fg.Show()
}

// OpenProjectLoadDialog opens a file dialog to select a project or EPUB file.
func (mw *MainWindow) OpenProjectLoadDialog(callback func(reader fyne.URIReadCloser, err error)) {
	fd := dialog.NewFileOpen(callback, mw.window)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{config.ProjectFileExt, ".epub"}))
	fd.Show()
}

// NewMainWindow creates a new instance of the main application window.
func NewMainWindow() error {
	if Main != nil {
		return errors.New("main window already exists")
	}
	Main = &MainWindow{
		fyneApp: app.NewWithID("org.saatool.app"),
		window:  nil,
		toolBar: container.NewHBox(),
		//header:  widget.NewLabel(fmt.Sprintf("SaaTool %v", config.Version)),
	}
	return nil
}

// ShowAndRun creates the main application window and starts the Fyne event loop.
func (mw *MainWindow) ShowAndRun() {
	defer log.Println("exiting application")

	err := config.LoadOptions()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	widgets.ApplyTheme(mw.fyneApp)

	mw.window = mw.fyneApp.NewWindow("SaaTool")
	mw.window.Resize(fyne.NewSize(800, 600))
	mw.window.SetMaster()

	mw.showTabs()
	mw.window.ShowAndRun()
}

// showTabs builds the three-tab main navigation and sets it as the window content.
// Must be called from the UI goroutine (or during setup before ShowAndRun).
func (mw *MainWindow) showTabs() {
	pv := NewProjectsView()
	sv := NewSettingsView()
	lv := NewLogView()

	mw.toolBar = container.NewHBox()

	// Theme toggle is always pinned to the right of the action bar.
	themeBtn := widget.NewButtonWithIcon("", widgets.IconTheme, mw.onThemeTapped)
	actionBar := container.NewBorder(nil, nil, nil, themeBtn, mw.toolBar)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Library", widgets.IconProject, pv.View()),
		container.NewTabItemWithIcon("Settings", widgets.IconSettings, sv.View()),
		container.NewTabItemWithIcon("Log", widgets.IconLog, lv.View()),
	)
	tabs.SetTabLocation(container.TabLocationBottom)

	views := []WindowContent{pv, sv, lv}
	tabs.OnChanged = func(_ *container.TabItem) {
		mw.ClearActions()
		idx := tabs.SelectedIndex()
		if idx >= 0 && idx < len(views) {
			views[idx].Load()
		}
	}

	layout := container.NewBorder(nil, actionBar, nil, nil, tabs)
	mw.window.SetContent(layout)
	pv.Load() // activate the Library tab immediately
}

// SetContent switches to fullscreen reading mode (e.g. TranslationView).
// Must be called from the UI goroutine.
func (mw *MainWindow) SetContent(content WindowContent) {
	if mw.reader != nil {
		mw.reader.Close()
	}
	mw.reader = content

	mw.toolBar = container.NewHBox()
	backBtn := widget.NewButtonWithIcon("Library", widgets.IconProject, mw.exitReader)
	themeBtn := widget.NewButtonWithIcon("", widgets.IconTheme, mw.onThemeTapped)
	topBar := container.NewBorder(nil, nil, backBtn, themeBtn)
	layout := container.NewBorder(topBar, nil, nil, nil, content.View())
	mw.window.SetContent(layout)
	content.Load()
}

// exitReader closes reading mode and returns to the library tabs.
func (mw *MainWindow) exitReader() {
	if mw.reader != nil {
		mw.reader.Close()
		mw.reader = nil
	}
	mw.showTabs()
}

// ApplyTheme applies the current theme and refreshes the UI.
func (mw *MainWindow) ApplyTheme() {
	widgets.ApplyTheme(mw.fyneApp)
	if mw.reader != nil {
		// In reading mode Fyne's SetTheme already triggers automatic repaints
		// for widget-based elements; canvas.Text objects need a manual nudge.
		mw.window.Canvas().Refresh(mw.window.Canvas().Content())
	} else {
		mw.showTabs()
	}
}

// Refresh is kept for compatibility.
func (mw *MainWindow) Refresh() {
	mw.ApplyTheme()
}

func (mw *MainWindow) ClearActions() {
	mw.toolBar.RemoveAll()
	mw.toolBar.Refresh()
}

func (mw *MainWindow) AddActionWidget(w fyne.CanvasObject) {
	mw.toolBar.Add(w)
	mw.toolBar.Refresh()
}

func (mw *MainWindow) AddAction(label string, icon fyne.Resource, action func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, action)
	mw.toolBar.Add(btn)
	mw.toolBar.Refresh()
	return btn
}

func (mw *MainWindow) onThemeTapped() {
	config.Options.DarkMode = !config.Options.DarkMode
	mw.ApplyTheme()
}

func (mw *MainWindow) ShowMessage(message string) {
	fyne.Do(dialog.NewInformation("Message", message, mw.window).Show)
}

func (mw *MainWindow) ShowError(message string) {
	fyne.Do(dialog.NewError(errors.New(message), mw.window).Show)
}

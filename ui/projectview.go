package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// ProjectView provides a UI for editing a translation.Project.
type ProjectView struct {
	project       *translation.Project
	nameEntry     *widget.Entry
	titleEntry    *widget.Entry
	authorEntry   *widget.Entry
	synopsisEntry *widget.Entry
	genreEntry    *widget.Entry
	promptEntry   *widget.Entry
	characterList *CharacterList
	View          fyne.CanvasObject
}

// NewProjectView creates a new ProjectEditor for the given project.
func NewProjectView() *ProjectView {
	pv := &ProjectView{}

	pv.nameEntry = widget.NewEntry()
	pv.nameEntry.Wrapping = fyne.TextWrapWord

	pv.titleEntry = widget.NewEntry()
	pv.authorEntry = widget.NewEntry()
	pv.synopsisEntry = widget.NewEntry()
	pv.synopsisEntry.Wrapping = fyne.TextWrapWord
	pv.synopsisEntry.MultiLine = true
	pv.synopsisEntry.SetMinRowsVisible(3)
	pv.genreEntry = widget.NewEntry()
	pv.promptEntry = widget.NewEntry()
	pv.promptEntry.Wrapping = fyne.TextWrapWord
	pv.promptEntry.MultiLine = true
	pv.promptEntry.SetMinRowsVisible(3)

	pv.characterList = NewCharactersList()

	characters := widgets.NewPanel(
		pv.characterList.View,
		fyne.NewSize(200, 100),
	)
	form := widget.NewForm(
		widget.NewFormItem("Name", pv.nameEntry),
		widget.NewFormItem("Title", pv.titleEntry),
		widget.NewFormItem("Author", pv.authorEntry),
		widget.NewFormItem("Genre", pv.genreEntry),
		widget.NewFormItem("Synopsis", pv.synopsisEntry),
		widget.NewFormItem("Prompt", pv.promptEntry),
		widget.NewFormItem("Characters", characters),
	)

	Main.ClearActions()
	Main.AddAction("Open", widgets.LoadIcon, pv.onOpenTapped)
	Main.AddAction("Translate", widgets.LoadIcon, pv.onTranslateTapped)

	pv.View = form

	pv.setActiveProject()

	return pv
}

// Save updates the Project fields from the UI entries.
func (ed *ProjectView) Save() {
	ed.project.Name = ed.nameEntry.Text
	ed.project.Title = ed.titleEntry.Text
	ed.project.Author = ed.authorEntry.Text
	ed.project.Synopsis = ed.synopsisEntry.Text
	ed.project.Genre = ed.genreEntry.Text
	ed.project.Prompt = ed.promptEntry.Text
}

// onOpenTapped handles the Open action for the project.
func (pl *ProjectView) onOpenTapped() {
	Main.OpenFileDialog(pl.onProjectFileOpened, ".json")
}

func (pl *ProjectView) onProjectFileOpened(reader fyne.URIReadCloser, err error) {
	defer pl.View.Refresh()

	if err != nil || reader == nil {
		return
	}
	defer reader.Close()
	projectPath := reader.URI().Path()
	pl.project, err = translation.LoadProject(projectPath)
	if err != nil {
		Main.ShowError(fmt.Sprintf("Failed to load project file '%v': %v", projectPath, err))
		return
	}
	pl.setProject(pl.project)
	config.Options.ActiveProjectFile = projectPath
	err = config.SaveSettings()
	if err != nil {
		Main.ShowError(fmt.Sprintf("Failed to save settings: %v", err))
		return
	}
}

// SetProject updates the ProjectCard with the given project details.
func (ed *ProjectView) setProject(project *translation.Project) {
	ed.project = project
	ed.nameEntry.SetText(project.Name)
	ed.titleEntry.SetText(project.Title)
	ed.authorEntry.SetText(project.Author)
	ed.synopsisEntry.SetText(project.Synopsis)
	ed.genreEntry.SetText(project.Genre)
	ed.promptEntry.SetText(project.Prompt)
	ed.characterList.SetCharacters(project.Characters)
}

func (ed *ProjectView) setProjectFromFile(filePath string) error {
	project, err := translation.LoadProject(filePath)
	if err != nil {
		return fmt.Errorf("failed to load project from file '%s': %w", filePath, err)
	}
	ed.setProject(project)
	return nil
}

func (ed *ProjectView) setActiveProject() {
	log.Printf("setting active project from file: '%s'", config.Options.ActiveProjectFile)
	if config.Options.ActiveProjectFile == "" {
		return
	}
	err := ed.setProjectFromFile(config.Options.ActiveProjectFile)
	if err != nil {
		log.Printf("Failed to set active project: %v", err)
		Main.ShowError(fmt.Sprintf("Failed to set active project: %v", err))
		return
	}
}

func (ed *ProjectView) onTranslateTapped() {
	if ed.project == nil {
		Main.ShowError("No project loaded to translate.")
		return
	}

	Main.SetContent(
		NewTranslationView(ed.project).View,
	)
}

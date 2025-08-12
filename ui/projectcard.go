package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// ProjectCard provides a UI for editing a translation.Project.
type ProjectCard struct {
	Project       *translation.Project
	nameEntry     *widget.Entry
	titleEntry    *widget.Entry
	authorEntry   *widget.Entry
	synopsisEntry *widget.Entry
	genreEntry    *widget.Entry
	promptEntry   *widget.Entry
	characterList *CharacterList
	View          fyne.CanvasObject
}

// SetProject updates the ProjectCard with the given project details.
func (ed *ProjectCard) SetProject(project *translation.Project) {
	ed.Project = project
	ed.nameEntry.SetText(project.Name)
	ed.titleEntry.SetText(project.Title)
	ed.authorEntry.SetText(project.Author)
	ed.synopsisEntry.SetText(project.Synopsis)
	ed.genreEntry.SetText(project.Genre)
	ed.promptEntry.SetText(project.Prompt)
	ed.characterList.SetCharacters(project.Characters)
}

// NewProjectEditor creates a new ProjectEditor for the given project.
func NewProjectEditor() *ProjectCard {
	ed := &ProjectCard{}

	ed.nameEntry = widget.NewEntry()
	ed.nameEntry.Wrapping = fyne.TextWrapWord

	ed.titleEntry = widget.NewEntry()
	ed.authorEntry = widget.NewEntry()
	ed.synopsisEntry = widget.NewEntry()
	ed.synopsisEntry.Wrapping = fyne.TextWrapWord
	ed.synopsisEntry.MultiLine = true
	ed.synopsisEntry.SetMinRowsVisible(3)
	ed.genreEntry = widget.NewEntry()
	ed.promptEntry = widget.NewEntry()
	ed.promptEntry.Wrapping = fyne.TextWrapWord
	ed.promptEntry.MultiLine = true
	ed.promptEntry.SetMinRowsVisible(3)

	ed.characterList = NewCharactersList()

	characters := widgets.NewPanel(
		ed.characterList.View,
		fyne.NewSize(200, 100),
	)
	form := widget.NewForm(
		widget.NewFormItem("Name", ed.nameEntry),
		widget.NewFormItem("Title", ed.titleEntry),
		widget.NewFormItem("Author", ed.authorEntry),
		widget.NewFormItem("Genre", ed.genreEntry),
		widget.NewFormItem("Synopsis", ed.synopsisEntry),
		widget.NewFormItem("Prompt", ed.promptEntry),
		widget.NewFormItem("Characters", characters),
	)

	card := widget.NewCard("Edit Project", "Edit the details of your translation project", form)

	ed.View = card
	return ed
}

// Save updates the Project fields from the UI entries.
func (ed *ProjectCard) Save() {
	ed.Project.Name = ed.nameEntry.Text
	ed.Project.Title = ed.titleEntry.Text
	ed.Project.Author = ed.authorEntry.Text
	ed.Project.Synopsis = ed.synopsisEntry.Text
	ed.Project.Genre = ed.genreEntry.Text
	ed.Project.Prompt = ed.promptEntry.Text
}

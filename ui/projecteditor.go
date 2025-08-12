package ui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
)

// ProjectEditor provides a UI for editing a translation.Project.
type ProjectEditor struct {
	Project       *translation.Project
	nameEntry     *widget.Entry
	titleEntry    *widget.Entry
	authorEntry   *widget.Entry
	synopsisEntry *widget.Entry
	genreEntry    *widget.Entry
	promptEntry   *widget.Entry
	charList      *widget.List
	addCharBtn    *widget.Button
	removeCharBtn *widget.Button
	sourceCount   *widget.Label
	targetCount   *widget.Label
	container     *fyne.Container
}

// NewProjectEditor creates a new ProjectEditor for the given project.
func NewProjectEditor(p *translation.Project) *ProjectEditor {
	ed := &ProjectEditor{Project: p}

	ed.nameEntry = widget.NewEntry()
	ed.nameEntry.SetText(p.Name)

	ed.titleEntry = widget.NewEntry()
	ed.titleEntry.SetText(p.Title)

	ed.authorEntry = widget.NewEntry()
	ed.authorEntry.SetText(p.Author)

	ed.synopsisEntry = widget.NewEntry()
	ed.synopsisEntry.SetText(p.Synopsis)

	ed.genreEntry = widget.NewEntry()
	ed.genreEntry.SetText(p.Genre)

	ed.promptEntry = widget.NewEntry()
	ed.promptEntry.SetText(p.Prompt)

	ed.sourceCount = widget.NewLabel(
		"Source paragraphs: " + strconv.Itoa(len(p.Source.Paragraphs)),
	)
	ed.targetCount = widget.NewLabel(
		"Target paragraphs: " + strconv.Itoa(len(p.Target.Paragraphs)),
	)

	ed.charList = widget.NewList(
		func() int { return len(p.Characters) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			label.SetText(p.Characters[i].Name)
		},
	)

	ed.addCharBtn = widget.NewButton("Add Character", func() {
	})

	ed.removeCharBtn = widget.NewButton("Remove Character", func() {
	})

	ed.container = container.NewVBox(
		widget.NewLabel("Edit Translation Project"),
		widget.NewForm(
			widget.NewFormItem("Name", ed.nameEntry),
			widget.NewFormItem("Title", ed.titleEntry),
			widget.NewFormItem("Author", ed.authorEntry),
			widget.NewFormItem("Synopsis", ed.synopsisEntry),
			widget.NewFormItem("Genre", ed.genreEntry),
			widget.NewFormItem("Prompt", ed.promptEntry),
		),
		ed.sourceCount,
		ed.targetCount,
		widget.NewLabel("Characters:"),
		container.NewHBox(ed.addCharBtn, ed.removeCharBtn),
		ed.charList,
	)

	return ed
}

// View returns the fyne.CanvasObject for this editor.
func (ed *ProjectEditor) View() fyne.CanvasObject {
	return ed.container
}

// Save updates the Project fields from the UI entries.
func (ed *ProjectEditor) Save() {
	ed.Project.Name = ed.nameEntry.Text
	ed.Project.Title = ed.titleEntry.Text
	ed.Project.Author = ed.authorEntry.Text
	ed.Project.Synopsis = ed.synopsisEntry.Text
	ed.Project.Genre = ed.genreEntry.Text
	ed.Project.Prompt = ed.promptEntry.Text
}

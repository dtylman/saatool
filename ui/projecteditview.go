package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/ai"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// ProjectEditView allows editing a project's metadata.
type ProjectEditView struct {
	project       *translation.Project
	view          fyne.CanvasObject
	entryTitle    *widget.Entry
	entryAuthor   *widget.Entry
	entrySynopsis *widget.Entry
	entryGenre    *widget.Entry
	entryFromLang *widget.Entry
	entryToLang   *widget.Entry
	styleSelector *widgets.StyleSelector
}

// NewProjectEditView creates a new view for editing project metadata.
func NewProjectEditView(project *translation.Project) *ProjectEditView {
	pv := &ProjectEditView{
		project:       project,
		entryTitle:    widget.NewEntry(),
		entryAuthor:   widget.NewEntry(),
		entrySynopsis: widget.NewMultiLineEntry(),
		entryGenre:    widget.NewEntry(),
		entryFromLang: widget.NewEntry(),
		entryToLang:   widget.NewEntry(),
	}

	pv.styleSelector = widgets.NewStyleSelector(ai.PromptStyle(project.Style), nil)

	pv.entryTitle.SetText(project.Title)
	pv.entryAuthor.SetText(project.Author)
	pv.entrySynopsis.SetText(project.Synopsis)
	pv.entryGenre.SetText(project.Genre)
	pv.entryFromLang.SetText(project.Source.Language)
	pv.entryToLang.SetText(project.Target.Language)

	pv.view = widget.NewForm(
		widget.NewFormItem("Title", pv.entryTitle),
		widget.NewFormItem("Author", pv.entryAuthor),
		widget.NewFormItem("Synopsis", pv.entrySynopsis),
		widget.NewFormItem("Genre", pv.entryGenre),
		widget.NewFormItem("Source Language", pv.entryFromLang),
		widget.NewFormItem("Target Language", pv.entryToLang),
		widget.NewFormItem("Translation Style", pv.styleSelector),
	)

	Main.ClearActions()
	Main.AddAction("Save", widgets.IconSave, pv.onSaveTapped)
	Main.AddAction("Cancel", widgets.IconPrev, pv.onCancelTapped)

	return pv
}

func (pv *ProjectEditView) View() fyne.CanvasObject {
	return pv.view
}

func (pv *ProjectEditView) Close() {}

func (pv *ProjectEditView) Load() {}

func (pv *ProjectEditView) onSaveTapped() {
	pv.project.Title = pv.entryTitle.Text
	pv.project.Author = pv.entryAuthor.Text
	pv.project.Synopsis = pv.entrySynopsis.Text
	pv.project.Genre = pv.entryGenre.Text
	pv.project.Source.Language = pv.entryFromLang.Text
	pv.project.Target.Language = pv.entryToLang.Text
	pv.project.Style = string(pv.styleSelector.SelectedStyle())

	_, err := pv.project.Save()
	if err != nil {
		log.Printf("failed to save project: %v", err)
		Main.ShowError(fmt.Sprintf("Failed to save project: %v", err))
		return
	}
	Main.ShowMessage("Project saved")
	Main.SetContent(NewProjectsView())
}

func (pv *ProjectEditView) onCancelTapped() {
	Main.SetContent(NewProjectsView())
}

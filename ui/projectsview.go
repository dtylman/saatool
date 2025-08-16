package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
)

// ProjectsView represents the project selection/editing UI.
type ProjectsView struct {
	View            fyne.CanvasObject
	projectEditor   *ProjectCard
	btnTranslate    *widget.Button
	selectedProject *translation.Project
}

// NewProjectsView creates a new ProjectsViewStruct instance.
func NewProjectsView() *ProjectsView {
	pv := &ProjectsView{}

	projectsList := NewProjectList()
	projectsList.OnSelected = pv.onProjectSelected

	pv.projectEditor = NewProjectEditor()

	pv.btnTranslate = widget.NewButton("Translate", pv.onTranslateTapped)
	pv.btnTranslate.Disable()

	Main.AddActionWidget(pv.btnTranslate)

	pv.View = container.NewVBox(
		projectsList.View, // top
		pv.projectEditor.View,
	)

	return pv
}

func (pv *ProjectsView) onTranslateTapped() {
	if pv.selectedProject == nil {
		Main.ShowMessage("Please select a project to translate.")
		return
	}

	Main.SetContent(NewTranslationView(pv.selectedProject).View)
}

// onProjectSelected handles the selection of a project from the list.
func (pv *ProjectsView) onProjectSelected(projectName string) {
	project, err := translation.LoadProjectFile(projectName)
	if err != nil {
		Main.ShowError(fmt.Sprintf("Failed to load project: %s", err.Error()))
		pv.projectEditor.SetProject(&translation.Project{})
		pv.selectedProject = nil
		return
	}
	pv.selectedProject = project
	pv.projectEditor.SetProject(project)
	pv.btnTranslate.Enable()
}

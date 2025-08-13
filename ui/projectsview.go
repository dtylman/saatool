package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
)

// ProjectsView represents the project selection/editing UI.
type ProjectsView struct {
	OnProjectSelected func(*translation.Project)
	View              fyne.CanvasObject
	projectEditor     *ProjectCard
	btnTranslate      *widget.Button
	selectedProject   *translation.Project
}

// NewProjectsView creates a new ProjectsViewStruct instance.
func NewProjectsView() *ProjectsView {
	pv := &ProjectsView{}

	projectsList := NewProjectList()
	projectsList.OnSelected = pv.onProjectSelected

	pv.projectEditor = NewProjectEditor()

	pv.btnTranslate = widget.NewButton("Translate", pv.onTranslateTapped)
	pv.btnTranslate.Disable()

	toolBar := container.NewVBox(pv.btnTranslate)
	pv.View = container.NewBorder(
		projectsList.View, // top
		toolBar,           //bottom
		nil,
		nil,
		container.NewVScroll(pv.projectEditor.View),
	)

	return pv
}

func (pv *ProjectsView) onTranslateTapped() {
	if pv.selectedProject == nil {
		dialog := widget.NewPopUp(
			widget.NewLabel("You must select a project first."),
			fyne.CurrentApp().Driver().AllWindows()[0].Canvas(),
		)
		dialog.Show()
		return
	}

	if pv.OnProjectSelected != nil {
		pv.OnProjectSelected(pv.selectedProject)
	}
}

// onProjectSelected handles the selection of a project from the list.
func (pv *ProjectsView) onProjectSelected(projectName string) {
	project, err := translation.LoadProjectFile(projectName)
	if err != nil {
		fyne.LogError("Failed to load project", err)
		pv.projectEditor.SetProject(&translation.Project{})
		pv.selectedProject = nil
		return
	}
	pv.selectedProject = project
	pv.projectEditor.SetProject(project)
	pv.btnTranslate.Enable()
}

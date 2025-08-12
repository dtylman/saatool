package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/dtylman/saatool/translation"
)

// ProjectsView represents the project selection/editing UI.
type ProjectsView struct {
	OnProjectSelected func()
	View              fyne.CanvasObject
	projectEditor     *ProjectEditor
}

// NewProjectsView creates a new ProjectsViewStruct instance.
func NewProjectsView() *ProjectsView {
	pv := &ProjectsView{}

	projectsList := NewProjectList()
	projectsList.OnSelected = pv.onProjectSelected

	pv.projectEditor = NewProjectEditor()

	pv.View = container.NewBorder(
		projectsList.View,
		container.NewVBox(),
		nil, nil,
		container.NewVScroll(pv.projectEditor.View),
	)

	return pv
}

// onProjectSelected handles the selection of a project from the list.
func (pv *ProjectsView) onProjectSelected(projectName string) {
	project, err := translation.LoadProjectFile(projectName)
	if err != nil {
		fyne.LogError("Failed to load project", err)
		pv.projectEditor.SetProject(&translation.Project{})
		return
	}

	pv.projectEditor.SetProject(project)

}

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
)

// ProjectsView represents the project selection/editing UI.
type ProjectsView struct {
	OnProjectSelected func()
	projectList       *widget.List
	selectBtn         *widget.Button
	createBtn         *widget.Button
	container         *fyne.Container
}

// NewProjectsView creates a new ProjectsViewStruct instance.
func NewProjectsView() *ProjectsView {
	pv := &ProjectsView{}

	pv.projectList = widget.NewList(
		func() int {
			return len(config.Projects.Projects)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Project Name")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			label.SetText("Project " + config.Projects.Projects[i])
		},
	)

	pv.selectBtn = widget.NewButton("Open Selected Project", func() {
		if pv.OnProjectSelected != nil {
			pv.OnProjectSelected()
		}
	})

	pv.createBtn = widget.NewButton("Create New Project", func() {
		// TODO: Show dialog to create a new project
	})

	pv.container = container.NewVBox(
		widget.NewLabel("Select or Create a Translation Project"),
		pv.projectList,
		pv.selectBtn,
		pv.createBtn,
	)

	return pv
}

// View returns the fyne.CanvasObject for this view.
func (pv *ProjectsView) View() fyne.CanvasObject {
	return pv.container
}

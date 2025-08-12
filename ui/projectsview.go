package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
)

// ProjectsView represents the project selection/editing UI.
type ProjectsView struct {
	OnProjectSelected func()
	projectList       *widget.List
	selectBtn         *widget.Button
	createBtn         *widget.Button
	editor            *ProjectEditor
	editorContainer   *fyne.Container
	container         *fyne.Container
	selectedIndex     int
}

// NewProjectsView creates a new ProjectsViewStruct instance.
func NewProjectsView() *ProjectsView {
	pv := &ProjectsView{}

	pv.selectedIndex = -1

	pv.projectList = widget.NewList(
		func() int {
			return len(config.Projects.Projects)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Project Item")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			project := config.Projects.Projects[id]
			item.(*widget.Label).SetText(project)
		},
	)

	pv.projectList.OnSelected = func(id widget.ListItemID) {
		pv.selectedIndex = int(id)
		pv.showEditor()
	}

	pv.selectBtn = widget.NewButton("Open Selected Project", func() {
		if pv.OnProjectSelected != nil {
			pv.OnProjectSelected()
		}
	})

	pv.createBtn = widget.NewButton("Create New Project", func() {
		// TODO: Show dialog to create a new project
	})

	pv.editorContainer = container.NewVBox() // Placeholder, will be filled by showEditor

	pv.container = container.NewVBox(
		widget.NewLabel("Select or Create a Translation Project"),
		pv.projectList,
		pv.selectBtn,
		pv.createBtn,
		pv.editorContainer,
	)

	return pv
}

// showEditor displays the ProjectEditor for the selected project
func (pv *ProjectsView) showEditor() {
	pv.editorContainer.Objects = nil

	if pv.selectedIndex < 0 || pv.selectedIndex >= len(config.Projects.Projects) {
		pv.editorContainer.Add(widget.NewLabel("No project selected"))
		return
	}
	project, err := translation.LoadProject(pv.selectedIndex)
	if err != nil {
		pv.editorContainer.Add(widget.NewLabel("Error loading project: " + err.Error()))
		return
	}
	pv.editor = NewProjectEditor(project)
	saveBtn := widget.NewButton("Save Project", func() {
		pv.editor.Save()
		// TODO: Save to file if needed
	})
	pv.editorContainer.Add(pv.editor.View())
	pv.editorContainer.Add(saveBtn)
	pv.editorContainer.Refresh()
	pv.editorContainer.Show()

}

// View returns the fyne.CanvasObject for this view.
func (pv *ProjectsView) View() fyne.CanvasObject {
	return pv.container
}

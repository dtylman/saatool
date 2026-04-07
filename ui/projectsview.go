package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// ProjectsView list the translation projects
type ProjectsView struct {
	selectedProject *translation.Project
	lstProjects     *widget.List
	view            fyne.CanvasObject
	projects        []config.ProjectFile
}

// NewProjectView creates a new ProjectEditor for the given project.
func NewProjectsView() *ProjectsView {
	pv := &ProjectsView{}

	Main.ClearActions()
	Main.AddAction("Translate", widgets.IconTranslate, pv.onTranslateTapped)
	Main.AddAction("Edit", widgets.IconSettings, pv.onEditTapped)
	Main.AddAction("Import", widgets.IconOpen, pv.onImportTapped)
	Main.AddAction("Export", widgets.IconSave, pv.onExportTapped)
	Main.AddAction("Delete", widgets.IconDelete, pv.onDeleteTapped)

	pv.lstProjects = widget.NewList(pv.lstProjectsLen, pv.lstProjectsCreateItem, pv.lstProjectsUpdateItem)
	pv.lstProjects.OnSelected = pv.onProjectSelected

	pv.listProjects()

	return pv
}

func (pl *ProjectsView) View() fyne.CanvasObject {
	return pl.view
}

func (pl *ProjectsView) Close() {
	// nothing to do
}
func (pl *ProjectsView) Load() {

}

func (pl *ProjectsView) setView() {
	if pl.lstProjects.Length() == 0 {
		pl.view = container.NewVBox(
			widget.NewLabel("No projects found."),
			widget.NewLabel("Use the Import button to add a project."),
		)
		return
	} else {
		pl.view = container.NewStack(pl.lstProjects)
	}
}

func (pl *ProjectsView) listProjects() {
	var err error
	pl.projects, err = config.ListProjects()
	if err != nil {
		log.Printf("failed to list projects: %v", err)
		return
	}
	pl.lstProjects.Refresh()
	pl.setView()
}

func (pl *ProjectsView) onProjectSelected(id widget.ListItemID) {
	if id < 0 || id >= len(pl.projects) {
		return
	}

	projectFile := pl.projects[id]

	project, err := translation.LoadProject(projectFile.Path)
	if err != nil {
		Main.ShowError(fmt.Sprintf("Failed to load project file '%s': %v", projectFile.Path, err))
		return
	}
	pl.setProject(project)

}

func (pl *ProjectsView) lstProjectsLen() int {
	return len(pl.projects)
}

func (pl *ProjectsView) lstProjectsCreateItem() fyne.CanvasObject {
	return widgets.NewListItem(widget.NewIcon(widgets.IconProject), "Project", "Not Loaded", nil)
}

func (pl *ProjectsView) lstProjectsUpdateItem(id widget.ListItemID, obj fyne.CanvasObject) {
	item := obj.(*widgets.ListItem)
	if id < 0 || id >= len(pl.projects) {
		log.Printf("invalid project id: %d", id)
		return
	}
	title := pl.projects[id].Name
	subTitle := "Not Loaded"
	if pl.selectedProject != nil && pl.selectedProject.Name == pl.projects[id].Name {
		subTitle = pl.selectedProject.Title
		item.SetSelected(true)
	} else {
		item.SetSelected(false)
	}

	item.SetTitle(title)
	item.SetSubtitle(subTitle)

}

// onExportTapped handles the Export action for the project.
func (pl *ProjectsView) onExportTapped() {
	if pl.selectedProject == nil {
		Main.ShowError("No project loaded to export.")
		return
	}
	Main.OpenProjectSaveDialog(pl.onProjectFileExported, pl.selectedProject)
}

func (pl *ProjectsView) onProjectFileExported(writer fyne.URIWriteCloser, err error) {
	if err != nil {
		Main.ShowError(fmt.Sprintf("Failed to export project file: %v", err))
		return
	}
	if writer == nil {
		Main.ShowError("No file selected to export project.")
		return
	}
	err1 := translation.ExportProject(pl.selectedProject, writer)
	if err1 != nil {
		Main.ShowError(fmt.Sprintf("Failed to export project file: %v", err1))
		return
	}
	writer.Close()
	Main.ShowMessage(fmt.Sprintf("Project exported to %s", writer.URI().Name()))
}

// onOpenTapped handles the Open action for the project.
func (pl *ProjectsView) onImportTapped() {
	Main.OpenProjectLoadDialog(pl.onProjectFileOpened)
}

func (pl *ProjectsView) onProjectFileOpened(reader fyne.URIReadCloser, err error) {
	fyne.Do(func() {

		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		projectPath, err := translation.ImportProject(reader)
		if err != nil {
			Main.ShowError(fmt.Sprintf("Failed to import project file: %v", err))
			return
		}
		log.Printf("imported project from %s", projectPath)
		pl.listProjects()
	})
}

// SetProject updates the ProjectCard with the given project details.
func (ed *ProjectsView) setProject(project *translation.Project) {
	ed.selectedProject = project
	ed.view.Refresh()
}

func (ed *ProjectsView) onTranslateTapped() {
	if ed.selectedProject == nil {
		Main.ShowError("No project loaded to translate.")
		return
	}

	tv, err := NewTranslationView(ed.selectedProject)
	if err != nil {
		Main.ShowError(fmt.Sprintf("Failed to create translation view: %v", err))
		return
	}
	Main.SetContent(tv)
}

func (ed *ProjectsView) onEditTapped() {
	if ed.selectedProject == nil {
		Main.ShowError("No project selected to edit.")
		return
	}
	Main.SetContent(NewProjectEditView(ed.selectedProject))
}

func (ed *ProjectsView) onDeleteTapped() {
	if ed.selectedProject == nil {
		Main.ShowError("No project loaded to delete.")
		return
	}

	msg := fmt.Sprintf("Are you sure you want to delete the project '%s'?", ed.selectedProject.Title)
	confirm := dialog.NewConfirm("Delete Project", msg, func(confirmed bool) {
		if !confirmed {
			return
		}
		err := translation.DeleteProject(ed.selectedProject)
		if err != nil {
			Main.ShowError(fmt.Sprintf("Failed to delete project: %v", err))
			return
		}
		ed.selectedProject = nil
		ed.listProjects()
		Main.SetContent(ed)
	}, Main.window)
	confirm.Show()
}

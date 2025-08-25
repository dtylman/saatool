package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/translation"
	"github.com/dtylman/saatool/ui/widgets"
)

// ProjectsView list the translation projects
type ProjectsView struct {
	selectedProject *translation.Project
	lstProjects     *widget.List
	View            fyne.CanvasObject
	projects        []config.ProjectFile
}

// NewProjectView creates a new ProjectEditor for the given project.
func NewProjectsView() *ProjectsView {
	pv := &ProjectsView{}

	Main.ClearActions()
	Main.AddAction("Import", widgets.IconOpen, pv.onImportTapped)
	Main.AddAction("Export", widgets.IconSave, pv.onExportTapped)
	Main.AddAction("Translate", widgets.IconTranslate, pv.onTranslateTapped)

	pv.lstProjects = widget.NewList(pv.lstProjectsLen, pv.lstProjectsCreateItem, pv.lstProjectsUpdateItem)
	pv.lstProjects.OnSelected = pv.onProjectSelected

	pv.listProjects()

	return pv
}

func (pl *ProjectsView) setView() {
	if pl.lstProjects.Length() == 0 {
		pl.View = container.NewVBox(
			widget.NewLabel("No projects found."),
			widget.NewLabel("Use the Import button to add a project."),
		)
		return
	} else {
		pl.View = container.NewStack(pl.lstProjects)
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
	//pl.lblSelectedProject.SetText(fmt.Sprintf("Selected Project: %s", projectFile.Name))

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
	return widgets.NewListItem(widget.NewIcon(widgets.IconProject), "Project", "Test Subtitle", nil, nil)
	// return widget.NewLabel("Project")
}

func (pl *ProjectsView) lstProjectsUpdateItem(id widget.ListItemID, obj fyne.CanvasObject) {
	item := obj.(*widgets.ListItem)
	if pl.selectedProject != nil {
		if pl.selectedProject.Name == pl.projects[id].Name {
			pl.lstProjects.Select(id)
			item.Title = fmt.Sprintf("%s (%s)", pl.projects[id].Name, pl.selectedProject.Title)
			item.Refresh()
			return
		}
	}
	if id < len(pl.projects) {
		item.Title = pl.projects[id].Name
		item.Subtitle = pl.projects[id].Path
		item.Refresh()

	}

}

// onExportTapped handles the Export action for the project.
func (pl *ProjectsView) onExportTapped() {
	if pl.selectedProject == nil {
		Main.ShowError("No project loaded to export.")
		return
	}
	// Main.SaveFileDialog(pl.project.Name+".json", pl.onProjectFileSaved)
}

// onOpenTapped handles the Open action for the project.
func (pl *ProjectsView) onImportTapped() {
	Main.OpenFileDialog(pl.onProjectFileOpened, config.ProjectFileExt)
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
	//ed.lblSelectedProject.SetText(fmt.Sprintf("Selected Project: %s", project.Name))
	ed.View.Refresh()
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
	Main.SetContent(tv.View)
}

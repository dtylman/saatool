package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/ui/widgets"
)

// ProjectList represents the project selection UI.
type ProjectList struct {
	// View is the main container for the project list UI.
	View       fyne.CanvasObject
	OnSelected func(project string)
}

// NewProjectList creates a new ProjectList instance with a card view.
func NewProjectList() *ProjectList {
	pl := &ProjectList{}

	list := widget.NewList(
		pl.len,
		pl.createItem,
		pl.updateItem,
	)
	list.OnSelected = pl.onSelected

	btnLoadProject := widget.NewToolbarAction(widgets.LoadIcon, pl.onLoadProject)
	toolbar := widget.NewToolbar(btnLoadProject)

	content := container.NewBorder(nil, nil, toolbar, nil, list)

	card := widget.NewCard("", "Projects", content)

	pl.View = widgets.NewPanel(card, fyne.NewSize(200, 200))
	return pl
}

func (pl *ProjectList) onLoadProject() {
	fd := dialog.NewFileOpen(pl.onProjectFileLoaded, fyne.CurrentApp().Driver().AllWindows()[0])
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	fd.Show()
}

func (pl *ProjectList) onProjectFileLoaded(reader fyne.URIReadCloser, err error) {
	if err != nil || reader == nil {
		return
	}
	defer reader.Close()
	projectPath := reader.URI().Path()
	config.Projects.Projects = append(config.Projects.Projects, projectPath)
	config.SaveProjects()
	pl.View.Refresh()
}

func (pl *ProjectList) len() int {
	return len(config.Projects.Projects)
}

// createItem creates a new item for the project list.
func (pl *ProjectList) createItem() fyne.CanvasObject {
	return widget.NewLabel("")
}

// updateItem updates the item at the specified index in the project list.
func (pl *ProjectList) updateItem(i int, item fyne.CanvasObject) {
	label := item.(*widget.Label)
	if i < len(config.Projects.Projects) {
		label.SetText(config.Projects.Projects[i])
	} else {
		label.SetText("Unknown Project")
	}
}

// onSelected handles the selection of an item in the project list.
func (pl *ProjectList) onSelected(id widget.ListItemID) {
	log.Printf("selected project: %v", id)
	if id < 0 || id >= len(config.Projects.Projects) {
		log.Println("Invalid project selection")
		return
	}
	project := config.Projects.Projects[id]
	log.Printf("Selected project: %s", project)
	if pl.OnSelected != nil {
		pl.OnSelected(project)
	}
}

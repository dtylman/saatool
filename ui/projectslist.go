package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/config"
	"github.com/dtylman/saatool/ui/widgets"
)

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

	card := widget.NewCard("Projects", "List of projects", list)

	pl.View = widgets.NewPanel(card, fyne.NewSize(200, 200))
	return pl
}

// len returns the number of items in the project list.
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

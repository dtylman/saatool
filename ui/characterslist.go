package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
)

// CharacterList represents the character selection UI.
type CharacterList struct {
	// Characters holds the list of characters in the project.
	Characters []translation.Character
	View       fyne.CanvasObject
}

// NewCharactersList creates a new CharactersList instance.
func NewCharactersList() *CharacterList {
	cl := &CharacterList{
		Characters: []translation.Character{},
	}

	cl.View = widget.NewList(cl.len, cl.createItem, cl.updateItem)
	return cl
}

// SetCharacters updates the CharacterList with the given characters.
func (cl *CharacterList) SetCharacters(characters []translation.Character) {
	cl.Characters = characters
	// cl.View.Objects = nil // Clear previous items
	// for _, character := range cl.Characters {
	// 	cl.View.Add(widget.NewLabel(character.Name))
	// 	log.Printf("Added character: %s", character.Name)
	// }

	cl.View.Refresh()
}

// len returns the number of characters in the list.
func (cl *CharacterList) len() int {
	return len(cl.Characters)
}

// createItem creates a new item for the character list.
func (cl *CharacterList) createItem() fyne.CanvasObject {
	return widget.NewLabel("")
}

// updateItem updates the item at the specified index in the character list.
func (cl *CharacterList) updateItem(i int, item fyne.CanvasObject) {
	label := item.(*widget.Label)
	if i < len(cl.Characters) {
		label.SetText(cl.Characters[i].Name)
	} else {
		label.SetText("Unknown Character")
	}
}

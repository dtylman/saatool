package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/saatool/translation"
)

// CharacterEditor provides a UI for editing a translation.Character.
type CharacterEditor struct {
	Character   *translation.Character
	nameEntry   *widget.Entry
	genderEntry *widget.Entry
	ageEntry    *widget.Entry
	roleEntry   *widget.Entry
	descEntry   *widget.Entry
	container   *fyne.Container
}

// NewCharacterEditor creates a new CharacterEditor for the given character.
func NewCharacterEditor(c *translation.Character) *CharacterEditor {
	ed := &CharacterEditor{Character: c}

	ed.nameEntry = widget.NewEntry()
	ed.nameEntry.SetText(c.Name)

	ed.genderEntry = widget.NewEntry()
	ed.genderEntry.SetText(c.Gender)

	ed.ageEntry = widget.NewEntry()
	ed.ageEntry.SetText(itoa(c.Age))

	ed.roleEntry = widget.NewEntry()
	ed.roleEntry.SetText(c.Role)

	ed.descEntry = widget.NewEntry()
	ed.descEntry.SetText(c.Description)

	ed.container = container.NewVBox(
		widget.NewLabel("Edit Character"),
		widget.NewForm(
			widget.NewFormItem("Name", ed.nameEntry),
			widget.NewFormItem("Gender", ed.genderEntry),
			widget.NewFormItem("Age", ed.ageEntry),
			widget.NewFormItem("Role", ed.roleEntry),
			widget.NewFormItem("Description", ed.descEntry),
		),
	)

	return ed
}

// View returns the fyne.CanvasObject for this editor.
func (ed *CharacterEditor) View() fyne.CanvasObject {
	return ed.container
}

// Save updates the Character fields from the UI entries.
func (ed *CharacterEditor) Save() {
	ed.Character.Name = ed.nameEntry.Text
	ed.Character.Gender = ed.genderEntry.Text
	ed.Character.Age = atoi(ed.ageEntry.Text)
	ed.Character.Role = ed.roleEntry.Text
	ed.Character.Description = ed.descEntry.Text
}

// Helper functions for int conversion
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func atoi(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

package widgets

import (
	"log"

	"fyne.io/fyne/v2/widget"
	"github.com/dtylman/aitasks/prompts"
)

// StyleSelector is a drop-down widget for selecting a translation prompt style.
type StyleSelector struct {
	widget.Select
}

// NewStyleSelector creates a new StyleSelector populated with the available styles.
// The onChange callback is called with the selected PromptStyle when the user picks one.
func NewStyleSelector(selected string, onChange func(string)) *StyleSelector {
	names, err := prompts.GetTranslateStyles()
	if err != nil {
		log.Printf("failed to get styles from prompts: %v", err)
		names = []string{}
	}
	s := &StyleSelector{}
	s.Options = names
	s.OnChanged = func(value string) {
		if onChange != nil {
			onChange(value)
		}
	}
	s.Selected = selected
	s.ExtendBaseWidget(s)
	return s
}

package widgets

import (
	"fyne.io/fyne/v2/widget"

	"github.com/dtylman/saatool/ai"
)

// StyleSelector is a drop-down widget for selecting a translation prompt style.
type StyleSelector struct {
	widget.Select
}

// NewStyleSelector creates a new StyleSelector populated with the available styles.
// The onChange callback is called with the selected PromptStyle when the user picks one.
func NewStyleSelector(selected ai.PromptStyle, onChange func(ai.PromptStyle)) *StyleSelector {
	names := ai.StyleNames()
	s := &StyleSelector{}
	s.Options = names
	s.OnChanged = func(value string) {
		if onChange != nil {
			onChange(ai.PromptStyle(value))
		}
	}
	s.Selected = string(selected)
	s.ExtendBaseWidget(s)
	return s
}

// SelectedStyle returns the currently selected PromptStyle.
func (s *StyleSelector) SelectedStyle() ai.PromptStyle {
	return ai.PromptStyle(s.Selected)
}

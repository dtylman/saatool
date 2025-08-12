package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Panel is a custom widget that wraps another widget and enforces a minimum size.
type Panel struct {
	widget.BaseWidget
	child   fyne.CanvasObject
	minSize fyne.Size
}

// NewPanel creates a new Panel with the specified child widget and minimum size.
func NewPanel(child fyne.CanvasObject, minSize fyne.Size) *Panel {
	panel := &Panel{
		child:   child,
		minSize: minSize,
	}
	panel.ExtendBaseWidget(panel)
	return panel
}

// CreateRenderer creates the renderer for our custom widget.
func (p *Panel) CreateRenderer() fyne.WidgetRenderer {
	return &panelRenderer{
		panel:   p,
		objects: []fyne.CanvasObject{p.child},
	}
}

// panelRenderer implements the WidgetRenderer interface.
type panelRenderer struct {
	panel   *Panel
	objects []fyne.CanvasObject
}

// MinSize returns the minimum size of the panel, as set in the constructor.
func (r *panelRenderer) MinSize() fyne.Size {
	return r.panel.minSize
}

// Layout arranges the child widget to fill the entire panel.
func (r *panelRenderer) Layout(size fyne.Size) {
	r.panel.child.Resize(size)
}

// Objects returns the objects to be rendered.
func (r *panelRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Refresh is called when the widget's state changes.
func (r *panelRenderer) Refresh() {
	r.panel.child.Refresh()
}

// Destroy is for cleaning up.
func (r *panelRenderer) Destroy() {}

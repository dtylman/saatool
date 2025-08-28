package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Panel is a custom widget that wraps another widget and enforces a minimum size.
type Panel struct {
	widget.BaseWidget
	child       fyne.CanvasObject
	rect        *canvas.Rectangle
	minSize     fyne.Size
	Border      float32
	BorderColor color.Color

	OnTapped func(pe *fyne.PointEvent) // Optional tap handler
}

// NewPanel creates a new Panel with the specified child widget and minimum size.
func NewPanel(child fyne.CanvasObject, minSize fyne.Size) *Panel {
	panel := &Panel{
		child:       child,
		minSize:     minSize,
		Border:      0,
		BorderColor: theme.Color(theme.ColorNameForeground),
	}
	if panel.Border > 0 {
		panel.rect = canvas.NewRectangle(panel.BorderColor)
		panel.rect.StrokeWidth = panel.Border
		panel.rect.FillColor = color.Transparent
	}
	panel.ExtendBaseWidget(panel)
	return panel
}

// CreateRenderer creates the renderer for our custom widget.
func (p *Panel) CreateRenderer() fyne.WidgetRenderer {
	objs := []fyne.CanvasObject{}
	if p.rect != nil {
		objs = append(objs, p.rect)
	}
	objs = append(objs, p.child)
	return &panelRenderer{
		panel:   p,
		objects: objs,
	}
}

// Tapped is called when the panel is tapped. It can be used to handle tap events.
func (p *Panel) Tapped(pe *fyne.PointEvent) {
	if p.OnTapped != nil {
		p.OnTapped(pe)
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
	if r.panel.child == nil {
		return
	}
	if r.panel.rect != nil {
		r.panel.rect.Resize(size)
	}
	r.panel.child.Resize(size)
}

// Objects returns the objects to be rendered.
func (r *panelRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Refresh is called when the widget's state changes.
func (r *panelRenderer) Refresh() {
	if r.panel.child == nil {
		return
	}
	if r.panel.rect != nil {
		r.panel.rect.FillColor = color.Transparent
		r.panel.rect.StrokeColor = r.panel.BorderColor
		r.panel.rect.StrokeWidth = r.panel.Border
	}
	r.panel.child.Refresh()
}

// Destroy is for cleaning up.
func (r *panelRenderer) Destroy() {}

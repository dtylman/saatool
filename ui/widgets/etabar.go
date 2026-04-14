package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ETABar is a simple progress bar that shows a colored rectangle proportional to its value.
type ETABar struct {
	widget.BaseWidget
	value float32
	Color color.Color
}

// NewETABar creates a new ETABar with a default semi-transparent blue color.
func NewETABar() *ETABar {
	b := &ETABar{
		Color: color.NRGBA{R: 70, G: 130, B: 220, A: 50},
	}
	b.ExtendBaseWidget(b)
	return b
}

// SetValue sets the progress value (0.0 to 1.0).
func (b *ETABar) SetValue(v float64) {
	b.value = float32(v)
	b.Refresh()
}

// CreateRenderer returns the renderer for the ETABar widget.
func (b *ETABar) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(b.Color)
	return &etaBarRenderer{bar: b, rect: rect}
}

type etaBarRenderer struct {
	bar  *ETABar
	rect *canvas.Rectangle
}

func (r *etaBarRenderer) Layout(size fyne.Size) {
	r.rect.Resize(fyne.NewSize(size.Width*r.bar.value, size.Height))
	r.rect.Move(fyne.NewPos(0, 0))
}

func (r *etaBarRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (r *etaBarRenderer) Refresh() {
	r.rect.FillColor = r.bar.Color
	size := r.bar.Size()
	if size.Width > 0 {
		r.rect.Resize(fyne.NewSize(size.Width*r.bar.value, size.Height))
	}
}

func (r *etaBarRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.rect}
}

func (r *etaBarRenderer) Destroy() {}

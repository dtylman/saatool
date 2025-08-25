package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ListItem is a custom widget similar to Flutter's ListTile.
type ListItem struct {
	widget.BaseWidget
	Leading  fyne.CanvasObject
	Title    string
	Subtitle string
	Trailing fyne.CanvasObject
	OnTapped func()
}

// NewListItem creates a new ListItem.
func NewListItem(leading fyne.CanvasObject, title, subtitle string, trailing fyne.CanvasObject, onTapped func()) *ListItem {
	item := &ListItem{
		Leading:  leading,
		Title:    title,
		Subtitle: subtitle,
		Trailing: trailing,
		OnTapped: onTapped,
	}
	item.ExtendBaseWidget(item)
	return item
}

func (i *ListItem) CreateRenderer() fyne.WidgetRenderer {
	title := canvas.NewText(i.Title, theme.ForegroundColor())
	title.TextStyle = fyne.TextStyle{Bold: true}
	var subtitle *canvas.Text
	if i.Subtitle != "" {
		subtitle = canvas.NewText(i.Subtitle, theme.DisabledTextColor())
	}
	texts := []fyne.CanvasObject{title}
	if subtitle != nil {
		texts = append(texts, subtitle)
	}
	textCol := container.NewVBox(texts...)

	objects := []fyne.CanvasObject{}
	if i.Leading != nil {
		objects = append(objects, i.Leading)
	}
	objects = append(objects, textCol)
	if i.Trailing != nil {
		objects = append(objects, i.Trailing)
	}

	row := container.NewHBox(objects...)
	bg := canvas.NewRectangle(theme.BackgroundColor())
	return &listItemRenderer{
		item:    i,
		bg:      bg,
		row:     row,
		objects: []fyne.CanvasObject{bg, row},
	}
}

func (i *ListItem) Tapped(_ *fyne.PointEvent) {
	if i.OnTapped != nil {
		i.OnTapped()
	}
}

func (i *ListItem) TappedSecondary(_ *fyne.PointEvent) {}

type listItemRenderer struct {
	item    *ListItem
	bg      *canvas.Rectangle
	row     *fyne.Container
	objects []fyne.CanvasObject
}

func (r *listItemRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	padding := theme.Padding()
	r.row.Move(fyne.NewPos(padding, padding))
	r.row.Resize(size.Subtract(fyne.NewSize(2*padding, 2*padding)))
}

func (r *listItemRenderer) MinSize() fyne.Size {
	return r.row.MinSize().Add(fyne.NewSize(2*theme.Padding(), 2*theme.Padding()))
}

func (r *listItemRenderer) Refresh() {
	r.bg.FillColor = theme.BackgroundColor()
	canvas.Refresh(r.bg)
	r.row.Refresh()
}

func (r *listItemRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *listItemRenderer) Destroy() {}

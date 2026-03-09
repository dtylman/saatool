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
	Leading     fyne.CanvasObject
	txtTitle    *canvas.Text
	txtSubTitle *canvas.Text
	Trailing    fyne.CanvasObject
	selected    bool
}

// NewListItem creates a new ListItem.
func NewListItem(leading fyne.CanvasObject, title, subtitle string, trailing fyne.CanvasObject) *ListItem {
	item := &ListItem{
		Leading:  leading,
		Trailing: trailing,
		selected: false,
	}

	item.txtTitle = canvas.NewText(title, theme.Color(theme.ColorNameForeground))
	item.txtTitle.TextStyle = fyne.TextStyle{Bold: true}
	item.txtTitle.TextSize = theme.Size(theme.SizeNameText) + 2

	item.txtSubTitle = canvas.NewText(subtitle, theme.Color(theme.ColorNamePlaceHolder))
	item.txtSubTitle.TextStyle = fyne.TextStyle{Italic: false}
	item.txtSubTitle.TextSize = theme.Size(theme.SizeNameText) - 2

	item.ExtendBaseWidget(item)
	return item
}

func (i *ListItem) SetTrailing(trailing fyne.CanvasObject) {
	i.Trailing = trailing
	i.Refresh()
}

func (i *ListItem) SetSelected(selected bool) {
	i.selected = selected
	i.Refresh()
}

func (i *ListItem) SetTitle(title string) {
	i.txtTitle.Text = title
	i.txtTitle.Refresh()
}

func (i *ListItem) SetSubtitle(subtitle string) {
	i.txtSubTitle.Text = subtitle
	i.txtSubTitle.Refresh()
}

func (i *ListItem) CreateRenderer() fyne.WidgetRenderer {
	texts := []fyne.CanvasObject{i.txtTitle, i.txtSubTitle}
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
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))

	return &listItemRenderer{
		item:    i,
		bg:      bg,
		row:     row,
		objects: []fyne.CanvasObject{bg, row},
	}
}

type listItemRenderer struct {
	item    *ListItem
	bg      *canvas.Rectangle
	row     *fyne.Container
	objects []fyne.CanvasObject
}

func (r *listItemRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	padding := theme.Padding() * 1.5
	r.row.Move(fyne.NewPos(padding, padding))
	r.row.Resize(size.Subtract(fyne.NewSize(2*padding, 2*padding)))
}

func (r *listItemRenderer) MinSize() fyne.Size {
	p := theme.Padding() * 1.5
	base := r.row.MinSize().Add(fyne.NewSize(2*p, 2*p))
	// enforce a comfortable minimum height
	if base.Height < 56 {
		base.Height = 56
	}
	return base
}

func (r *listItemRenderer) Refresh() {
	if r.item.selected {
		r.bg.FillColor = theme.Color(theme.ColorNameSelection)
	} else {
		r.bg.FillColor = theme.Color(theme.ColorNameBackground)
	}
	// re-sync text colors in case theme changed
	r.item.txtTitle.Color = theme.Color(theme.ColorNameForeground)
	r.item.txtSubTitle.Color = theme.Color(theme.ColorNamePlaceHolder)
	canvas.Refresh(r.bg)
	r.row.Refresh()
}

func (r *listItemRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *listItemRenderer) Destroy() {}

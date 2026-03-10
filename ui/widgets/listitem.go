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
		Leading:     leading,
		txtTitle:    canvas.NewText(title, theme.Color(theme.ColorNameForeground)),
		txtSubTitle: canvas.NewText(subtitle, theme.Color(theme.ColorNameForeground)),
		Trailing:    trailing,
		selected:    false,
	}
	item.txtTitle.TextStyle = fyne.TextStyle{Bold: true}
	item.txtTitle.TextSize = theme.Size(theme.SizeNameText)

	item.txtSubTitle = canvas.NewText(subtitle, theme.Color(theme.ColorNameForeground))
	item.txtSubTitle.TextStyle = fyne.TextStyle{Italic: true}
	subtitleSize := theme.Size(theme.SizeNameText) - 4
	if subtitleSize < 0 {
		subtitleSize = 1
	}
	item.txtSubTitle.TextSize = subtitleSize

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
	if i.selected {
		bg.FillColor = theme.Color(theme.ColorNameSelection)
	}
	return &listItemRenderer{
		item:    i,
		bg:      bg,
		row:     row,
		objects: []fyne.CanvasObject{bg, row},
	}
}

// func (i *ListItem) Tapped(_ *fyne.PointEvent) {
// 	log.Println("ListItem tapped")
// 	if i.OnTapped != nil {
// 		i.OnTapped()
// 	}
// }

// func (i *ListItem) TappedSecondary(_ *fyne.PointEvent) {}

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
	if r.item.selected {
		r.bg.FillColor = theme.Color(theme.ColorNameSelection)
	} else {
		r.bg.FillColor = theme.Color(theme.ColorNameBackground)
	}
	canvas.Refresh(r.bg)
	r.row.Refresh()
}

func (r *listItemRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *listItemRenderer) Destroy() {}

package widgets

import (
	"image/color"
	"log"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Direction defines the text direction for BiDi text rendering.
type Direction int

const (
	// LeftToRight indicates that the text should be rendered from left to right.
	LeftToRight Direction = iota
	// RightToLeft indicates that the text should be rendered from right to left.
	RightToLeft
)

// BidiText is a widget that renders BiDi text with support for RTL and LTR directions.
type BidiText struct {
	widget.BaseWidget
	// Words contains the words to be rendered in BiDi text.
	Words []string
	// Color is the color of the text.
	Color color.Color
	// Direction indicates the text direction (LeftToRight or RightToLeft).
	Direction Direction
	// Spacing is the space between words.
	Spacing float32
	// TextSize is the size of the text.
	TextSize float32
	// Padding is the padding around the text.
	Padding float32
}

func NewBidiText(text string) *BidiText {
	text = strings.Replace(text, "\n", " <NL> ", -1)
	b := &BidiText{
		Words:     strings.Fields(text),
		Direction: RightToLeft,
		TextSize:  theme.TextSize(),
		Padding:   theme.InnerPadding(),
		Spacing:   6,
		Color:     theme.Color(theme.ColorNameForeground),
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *BidiText) CreateRenderer() fyne.WidgetRenderer {
	r := &bidiTextRenderer{
		parent: b,
	}

	r.rect = canvas.NewRectangle(color.Transparent)
	r.rect.StrokeWidth = 1
	r.rect.StrokeColor = b.Color
	r.rect.FillColor = color.Transparent

	r.texts = make([]*canvas.Text, 0)

	var words = b.Words
	if b.Direction == RightToLeft {
		slices.Reverse(words)
		// for i := range words {
		// 	words[i] = reversePunctuation(words[i])
		// }
	}

	for _, word := range words {
		text := canvas.NewText(word, b.Color)
		text.TextSize = b.TextSize
		r.texts = append(r.texts, text)
	}

	r.initializeObjects()

	return r
}

func reversePunctuation(s string) string {
	punctuations := []string{".", ",", "!", "?"}
	for _, p := range punctuations {
		if strings.HasSuffix(s, p) {
			log.Printf("Reversing punctuation in: %s", s)
			return p + strings.TrimSuffix(s, p)
		}
	}
	return s
}

type bidiTextRenderer struct {
	parent  *BidiText
	rect    *canvas.Rectangle
	texts   []*canvas.Text
	objects []fyne.CanvasObject
}

func (r *bidiTextRenderer) initializeObjects() {
	r.objects = make([]fyne.CanvasObject, len(r.texts)+1)
	r.objects[0] = r.rect

	for i, text := range r.texts {
		r.objects[i+1] = text
		text.Alignment = fyne.TextAlignCenter
	}
}

func (r *bidiTextRenderer) Layout(size fyne.Size) {
	r.rect.Resize(size)
	if r.parent.Direction == LeftToRight {
		r.layoutLTR(size)
	} else {
		r.layoutRTL(size)
	}
}

func (r *bidiTextRenderer) layoutRTL(size fyne.Size) {
	x := size.Width - r.parent.Padding
	y := r.parent.Padding
	for i := len(r.texts) - 1; i >= 0; i-- {
		text := r.texts[i]
		log.Printf(text.Text)
		if text.Text == "<NL>" {
			// If the text is a newline, just move down without resizing
			log.Printf("Newline detected, moving down")
			y += text.MinSize().Height + r.parent.Spacing
			x = size.Width - r.parent.Padding // Reset x to padding for the next line
			continue
		}
		width := text.MinSize().Width
		if x-width < r.parent.Padding {
			x = size.Width - r.parent.Padding             // Reset x to padding if it exceeds width
			y += text.MinSize().Height + r.parent.Spacing // Move to next line
		}
		text.Resize(fyne.NewSize(width, text.MinSize().Height))
		text.TextSize = r.parent.TextSize
		text.Alignment = fyne.TextAlignCenter
		text.Move(fyne.NewPos(x-width, y))
		x -= width + r.parent.Spacing // Move x to the left for the next text
	}
}

func (r *bidiTextRenderer) layoutLTR(size fyne.Size) {
	x := r.parent.Padding
	y := r.parent.Padding
	for _, text := range r.texts {
		width := text.MinSize().Width
		if x+width > size.Width-r.parent.Padding {
			x = r.parent.Padding                          // Reset x to padding if it exceeds width
			y += text.MinSize().Height + r.parent.Spacing // Move to next line
		}
		text.Resize(fyne.NewSize(width, text.MinSize().Height))
		text.TextSize = r.parent.TextSize
		text.Alignment = fyne.TextAlignCenter
		text.Move(fyne.NewPos(x, y))
		x += text.MinSize().Width + r.parent.Spacing // Move x to the right for the next text
	}
}

func (r *bidiTextRenderer) MinSize() fyne.Size {
	return fyne.NewSize(40, 40)
}

func (r *bidiTextRenderer) Refresh() {
	r.rect.Refresh()
}

func (r *bidiTextRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *bidiTextRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *bidiTextRenderer) Destroy() {}

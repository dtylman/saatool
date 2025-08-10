package main

import (
	"image/color"
	"log"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

var sampleText = "יום יום אני תולש מהלוח דף. יום ראשון - כמעט. יום שני - I'm happy! ויום שלישי - 365 ימים בשנה!"

//var sampleText = "Hello World! This is a sample text to demonstrate the bidirectional text rendering in Fyne."

type Direction int

const (
	LeftToRight Direction = iota
	RightToLeft
)

type BidiText struct {
	widget.BaseWidget
	Words     []string
	Border    color.Color
	Direction Direction
	Spacing   float32
	TextSize  float32
	Padding   float32
}

func NewBidiText(text string) *BidiText {
	b := &BidiText{
		Words:     strings.Fields(text),
		Border:    color.Black,
		Direction: RightToLeft,
		TextSize:  16,
		Padding:   10,
		Spacing:   6,
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *BidiText) CreateRenderer() fyne.WidgetRenderer {
	r := &bidiTextRenderer{
		parent: b,
	}

	r.rect = canvas.NewRectangle(color.Transparent)
	r.rect.StrokeColor = b.Border
	r.rect.StrokeWidth = 2
	r.rect.FillColor = color.Transparent

	r.texts = make([]*canvas.Text, 0)

	var words = b.Words
	if b.Direction == RightToLeft {
		slices.Reverse(words)
		for i := range words {
			words[i] = reversePunctuation(words[i])
		}
	}

	for _, word := range words {
		text := canvas.NewText(word, color.Black)
		log.Printf("Word %v, MinSize: %v Size: %v", word, text.MinSize(), text.Size())
		r.texts = append(r.texts, text)
	}

	r.initializeObjects()

	return r
}

func reversePunctuation(s string) string {
	reversed := make([]rune, len(s))
	for i, r := range s {
		if r == '!' || r == '?' || r == '.' {
			reversed[len(s)-1-i] = r
		} else {
			reversed[i] = r
		}
	}
	return string(reversed)
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
	} else {
		x := size.Width - r.parent.Padding
		y := r.parent.Padding
		for i := len(r.texts) - 1; i >= 0; i-- {
			text := r.texts[i]
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

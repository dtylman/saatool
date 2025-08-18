package widgets

import (
	"image/color"
	"log"
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
	// TextStyle is the style of the text.
	TextStyle fyne.TextStyle
	// Padding is the padding around the text.
	Padding float32
	// Offset is the index of the first word currently visible in the widget.
	Offset int
	// Length is the index of the last word currently visible in the widget.
	Length int
}

// NewBidiText creates a new BidiText widget with default settings.
func NewBidiText() *BidiText {
	b := &BidiText{
		Words:     []string{},
		Direction: RightToLeft,
		TextSize:  theme.TextSize(),
		TextStyle: fyne.TextStyle{},
		Padding:   theme.InnerPadding(),
		Spacing:   theme.TextSize(),
		Color:     theme.Color(theme.ColorNameForeground),
		Offset:    0,
		Length:    0,
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *BidiText) SetDirection(direction Direction) {
	if b.Direction != direction {
		b.Direction = direction
		b.Refresh()
	}
}

func (b *BidiText) SetWords(words []string) {
	b.Words = words
	b.Offset = 0
	b.Length = len(words)
	if b.Direction == RightToLeft {
		// slices.Reverse(b.Words)
		// for i := range words {
		// 	words[i] = reversePunctuation(words[i])
		// }
	}
	b.Refresh()
}

// SetColor sets the color of the BidiText.
func (b *BidiText) SetColor(color color.Color) {
	if b.Color != color {
		b.Color = color
		b.Refresh()
	}
}

// SetTextSize sets the size of the text in the BidiText widget.
func (b *BidiText) SetTextSize(size float32) {
	if b.TextSize != size {
		b.TextSize = size
		b.Refresh()
	}
}

func (b *BidiText) Next() bool {
	if (b.Offset + b.Length) < len(b.Words) {
		b.Offset += b.Length
		b.Refresh()
		return true // More words to show
	}
	return false // End of paragraph
}

func (b *BidiText) Previous() bool {
	if b.Offset == 0 { //already at the beginning
		return false
	}
	start := 0
	end := b.Offset

	for start < end {
		previousWords := b.Words[start:end]
		visibleCount := b.measureLayout(previousWords)
		if len(previousWords) == visibleCount {
			break
		}
		start++
	}

	b.Offset = start
	b.Refresh()
	return true
}

func (r *BidiText) isNewLine(text string) bool {
	return strings.TrimSpace(text) == "<NL>"
}

// measureLayout BidiText how many words can fit from the provided list
func (r *BidiText) measureLayout(words []string) int {
	lineHeight := r.TextSize + r.Spacing
	minX := r.Padding
	minY := r.Padding
	maxX := r.Size().Width - r.Padding
	maxY := r.Size().Height - (r.Padding * 2)

	x := minX
	y := minY
	for i, word := range words {
		width := fyne.MeasureText(word, r.TextSize, r.TextStyle).Width
		if r.isNewLine(word) { // If the text is a newline, just move down without resizing
			y += lineHeight
			x = minX
			continue
		}

		x += width + r.Spacing
		if x >= maxX {
			x = minX        // Reset x to padding if it exceeds width
			y += lineHeight // Move to next line
		}

		if y+lineHeight >= maxY { // Reached the end of the available space
			return i
		}

	}
	return len(words) // all words fit
}

// NewBidiTextWithWords creates a new BidiText widget with the specified words and default settings.
func (b *BidiText) CreateRenderer() fyne.WidgetRenderer {
	r := &bidiTextRenderer{
		parent: b,
	}

	r.rect = canvas.NewRectangle(color.Transparent)
	r.rect.StrokeWidth = 1
	r.rect.StrokeColor = b.Color
	r.rect.FillColor = color.Transparent

	r.texts = make([]*canvas.Text, 0)

	r.initializeObjects()

	return r
}

func reversePunctuation(s string) string {
	punctuations := []string{".", ",", "!", "?", ":"}
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
	size    fyne.Size
}

func (r *bidiTextRenderer) initializeObjects() {
	r.texts = make([]*canvas.Text, 0)

	if r.parent.Offset >= len(r.parent.Words) {
		log.Printf("offset %d exceeds words length %d", r.parent.Offset, len(r.parent.Words))
		return
	}

	for i := r.parent.Offset; i < len(r.parent.Words); i++ {
		word := r.parent.Words[i]
		text := canvas.NewText(word, r.parent.Color)
		text.TextSize = r.parent.TextSize
		text.Alignment = fyne.TextAlignCenter
		r.texts = append(r.texts, text)
	}

	r.objects = make([]fyne.CanvasObject, len(r.texts)+1)
	r.objects = append(r.objects, r.rect) // Add the rectangle as the first object
	for i, text := range r.texts {
		r.objects[i+1] = text
	}
}

func (r *bidiTextRenderer) Layout(size fyne.Size) {
	r.rect.Resize(size)
	r.size = size
	r.drawTexts()
}

func (r *bidiTextRenderer) drawTexts() {
	if r.parent.Direction == LeftToRight {
		r.layoutLTR()
	} else {
		r.layoutRTL()
	}
}

func (r *bidiTextRenderer) layoutRTL() {
	lineHeight := r.parent.TextSize + r.parent.Spacing
	minX := r.parent.Padding
	minY := r.parent.Padding
	maxX := r.size.Width - (r.parent.Padding)
	maxY := r.size.Height - (r.parent.Padding * 2)

	x := maxX
	y := minY

	for i, text := range r.texts {
		text.Show()                        // make sure all texts are visible
		if r.parent.isNewLine(text.Text) { // If the text is a newline, just move down without resizing
			y += lineHeight
			x = maxX
			continue
		}

		textWidth := text.MinSize().Width // get the width of the text

		if x-textWidth-r.parent.Spacing < minX { // Check if x exceeds the minimum x
			x = maxX
			y += lineHeight // Move to next line
		}

		if y > maxY-lineHeight { // Check if y exceeds the height of the rectangle
			// hide the remaining texts
			for j := i; j < len(r.texts); j++ {
				r.texts[j].Hide()
			}
			r.parent.Length = i // Update visible words count
			return
		}

		x -= (textWidth + r.parent.Spacing) // Move x to the right for the next text
		text.Resize(fyne.NewSize(textWidth, lineHeight))
		text.TextSize = r.parent.TextSize
		text.Alignment = fyne.TextAlignTrailing
		text.Move(fyne.NewPos(x, y))

	}
	r.parent.Length = len(r.texts) // Update visible words count
}

func (r *bidiTextRenderer) layoutLTR() {
	lineHeight := r.parent.TextSize + r.parent.Spacing
	minX := r.parent.Padding
	minY := r.parent.Padding
	maxX := r.size.Width - (r.parent.Padding * 2)
	maxY := r.size.Height - (r.parent.Padding * 2)

	x := minX
	y := minY

	for i, text := range r.texts {
		text.Show()                        // make sure all texts are visible
		if r.parent.isNewLine(text.Text) { // If the text is a newline, just move down without resizing
			y += lineHeight
			x = minX
			continue
		}

		width := text.MinSize().Width // get the width of the text

		if x+width > maxX {
			x = minX        // Reset x to padding if it exceeds width
			y += lineHeight // Move to next line
		}

		if y > maxY-lineHeight { // Check if y exceeds the height of the rectangle
			// hide the remaining texts
			for j := i; j < len(r.texts); j++ {
				r.texts[j].Hide()
			}
			r.parent.Length = i // Update visible words count
			return
		}

		text.Resize(fyne.NewSize(width, text.MinSize().Height))
		text.TextSize = r.parent.TextSize
		text.Alignment = fyne.TextAlignCenter
		text.Move(fyne.NewPos(x, y))
		x += text.MinSize().Width + r.parent.Spacing // Move x to the right for the next text
	}
	r.parent.Length = len(r.texts) // Update visible words count
}

func (r *bidiTextRenderer) MinSize() fyne.Size {
	return fyne.NewSize(40, 40)
}

func (r *bidiTextRenderer) Refresh() {
	r.initializeObjects()
	r.rect.Refresh()

}

func (r *bidiTextRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *bidiTextRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *bidiTextRenderer) Destroy() {}

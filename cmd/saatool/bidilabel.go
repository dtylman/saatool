package main

import (
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BidiLabel is a custom read-only widget for better bidirectional text support.
// It will manually split text into RTL and LTR segments and position them correctly.
type BidiLabel struct {
	widget.BaseWidget
	Text      string
	textStyle fyne.TextStyle
}

// NewBidiLabel creates a new BidiLabel widget.
func NewBidiLabel(text string) *BidiLabel {
	l := &BidiLabel{
		Text:      text,
		textStyle: fyne.TextStyle{},
	}
	l.ExtendBaseWidget(l)
	return l
}

// CreateRenderer is a required method for a custom widget. It returns the renderer
// that will draw the widget's content to the canvas.
func (l *BidiLabel) CreateRenderer() fyne.WidgetRenderer {
	return &bidiLabelRenderer{
		bidiLabel: l,
	}
}

// bidiLabelRenderer is the renderer for our custom BidiLabel widget.
type bidiLabelRenderer struct {
	bidiLabel    *BidiLabel
	objects      []fyne.CanvasObject
	wrappedLines []string
}

// Objects returns the canvas objects that make up the widget.
func (r *bidiLabelRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy is called when the widget is removed from the canvas.
func (r *bidiLabelRenderer) Destroy() {}

// MinSize calculates the minimum size needed to display the text.
func (r *bidiLabelRenderer) MinSize() fyne.Size {
	// Properly handle explicit newlines for minimum size calculation
	r.wrappedLines = r.wrapText(r.bidiLabel.Text, fyne.NewSize(0, 0))
	lineHeight := fyne.MeasureText("M", theme.TextSize(), r.bidiLabel.textStyle).Height
	maxWidth := float32(0)
	for _, line := range r.wrappedLines {
		w := fyne.MeasureText(line, theme.TextSize(), r.bidiLabel.textStyle).Width
		if w > maxWidth {
			maxWidth = w
		}
	}
	return fyne.NewSize(maxWidth, float32(len(r.wrappedLines))*lineHeight)
}

// Refresh is called when the widget's content needs to be redrawn.
func (r *bidiLabelRenderer) Refresh() {
	r.Layout(r.bidiLabel.Size())
	canvas.Refresh(r.bidiLabel)
}

// Layout arranges the text segments in the correct bidirectional order.
func (r *bidiLabelRenderer) Layout(size fyne.Size) {
	// Clear old objects
	r.objects = nil

	// Wrap the text into lines based on the new size
	r.wrappedLines = r.wrapText(r.bidiLabel.Text, size)

	currentY := float32(0)
	lineHeight := fyne.MeasureText("M", theme.TextSize(), r.bidiLabel.textStyle).Height

	for _, line := range r.wrappedLines {
		// Support explicit newlines: if the line is "", just increment Y
		if line == "" {
			currentY += lineHeight
			continue
		}
		segments := splitTextByDirection(line)
		currentX := size.Width
		for _, s := range segments {
			t := canvas.NewText(s, theme.ForegroundColor())
			t.TextStyle = r.bidiLabel.textStyle
			t.TextSize = theme.TextSize()
			textSize := fyne.MeasureText(s, t.TextSize, t.TextStyle)
			currentX -= textSize.Width
			t.Move(fyne.NewPos(currentX, currentY))
			r.objects = append(r.objects, t)
		}
		currentY += lineHeight
	}
}

// wrapText breaks the given text into lines that fit within the specified width, supporting explicit '\n'.
func (r *bidiLabelRenderer) wrapText(text string, size fyne.Size) []string {
	// Split on newlines first
	rawLines := strings.Split(text, "\n")
	var lines []string
	for _, rawLine := range rawLines {
		words := strings.Split(rawLine, " ")
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := ""
		for _, word := range words {
			// Check if adding the next word would exceed the width
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			testSize := fyne.MeasureText(testLine, theme.TextSize(), r.bidiLabel.textStyle)

			if size.Width > 0 && testSize.Width > size.Width && currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				if currentLine != "" {
					currentLine += " "
				}
				currentLine += word
			}
		}
		lines = append(lines, currentLine)
	}
	return lines
}

// splitTextByDirection tokenizes the string into segments of consecutive
// Hebrew (RTL) or non-Hebrew (LTR) characters.
func splitTextByDirection(text string) []string {
	var segments []string
	var currentSegment string
	var isRTL bool // Assume LTR for the first character

	for i, r := range text {
		isHebrew := isHebrewChar(r)

		if i == 0 {
			isRTL = isHebrew
		}

		// Check if the directionality has changed.
		if isHebrew != isRTL && currentSegment != "" {
			segments = append(segments, currentSegment)
			currentSegment = string(r)
			isRTL = isHebrew
		} else {
			currentSegment += string(r)
		}
	}
	// Add the last segment
	if currentSegment != "" {
		segments = append(segments, currentSegment)
	}

	// Only reverse the order of segments if the line starts with Hebrew (RTL)
	if len(segments) > 1 && isHebrewChar(rune(text[0])) {
		for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
			segments[i], segments[j] = segments[j], segments[i]
		}
	}

	return segments
}

// isHebrewChar checks if a rune is a Hebrew character.
func isHebrewChar(r rune) bool {
	return unicode.Is(unicode.Hebrew, r)
}

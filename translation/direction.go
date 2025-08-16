package translation

import (
	"log"

	iso6391 "github.com/emvi/iso-639-1"
)

// Direction represents the text direction for rendering.
type Direction int

const (
	// LeftToRight indicates that the text should be rendered from left to right.
	LeftToRight Direction = iota
	// RightToLeft indicates that the text should be rendered from right to left.
	RightToLeft
)

// GetTextDirection returns the text direction for the given language code.
func GetTextDirection(lang string) Direction {
	code := iso6391.CodeForName(lang)
	if code == "" {
		code = iso6391.CodeForNativeName(lang)
		if code == "" {
			code = iso6391.Name(lang)
		}
	}
	log.Printf("rtl: code for '%v' is '%v'", lang, code)
	if code == "ar" || code == "he" || code == "fa" || code == "ur" {
		return RightToLeft
	}

	return LeftToRight
}

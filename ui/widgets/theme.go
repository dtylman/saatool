package widgets

import (
	_ "embed"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed SimpleCLM-Medium.ttf
var SimpleCLMMedium []byte

// Theme implements a custom theme for the application.
type Theme struct {
	baseSize int
}

// NewTheme creates a new instance of the custom theme with the specified base size.
func NewTheme(baseSize int) *Theme {
	return &Theme{
		baseSize: baseSize,
	}
}

// Color returns the color for the specified theme color name and variant.
func (t *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

// Font returns the font resource for the specified text style.
func (t *Theme) Font(s fyne.TextStyle) fyne.Resource {
	return fyne.NewStaticResource("SimpleCLM-Medium.ttf", SimpleCLMMedium)
}

// Icon returns the icon resource for the specified theme icon name.
func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the specified theme size name.
func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case fyne.ThemeSizeName("innerPadding"):
		return 3
	case fyne.ThemeSizeName("iconInline"):
		return float32(t.baseSize * 2)
	case fyne.ThemeSizeName("text"):
		return float32(t.baseSize)
	case fyne.ThemeSizeName("padding"):
		return float32(t.baseSize / 2)
	case fyne.ThemeSizeName("scrollBar"):
		return float32(t.baseSize * 2)
	case fyne.ThemeSizeName("scrollBarSmall"):
		return float32(t.baseSize)
	case fyne.ThemeSizeName("scrollBarRadius"):
		return float32(t.baseSize / 2)
	case fyne.ThemeSizeName("separator"):
		return 1
	case fyne.ThemeSizeName("lineSpacing"):
		return 1.5
	case fyne.ThemeSizeName("selectionRadius"):
		return 4
	case fyne.ThemeSizeName("inputBorder"):
		return 2
	case fyne.ThemeSizeName("inputRadius"):
		return 4

	default:
		log.Println(name)
	}

	return theme.DefaultTheme().Size(name)
}

package widgets

import (
	_ "embed"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/dtylman/saatool/config"
)

//go:embed SimpleCLM-Medium.ttf
var SimpleCLMMedium []byte

// palette entries for dark and light modes
var (
	darkBackground  = color.NRGBA{R: 0x1e, G: 0x1e, B: 0x2e, A: 0xff}
	darkForeground  = color.NRGBA{R: 0xcd, G: 0xd6, B: 0xf4, A: 0xff}
	darkButton      = color.NRGBA{R: 0x2a, G: 0x2a, B: 0x3e, A: 0xff}
	darkAccent      = color.NRGBA{R: 0xcb, G: 0xa6, B: 0xf7, A: 0xff}
	darkInputBg     = color.NRGBA{R: 0x31, G: 0x32, B: 0x44, A: 0xff}
	darkPlaceholder = color.NRGBA{R: 0x58, G: 0x5b, B: 0x70, A: 0xff}

	lightBackground  = color.NRGBA{R: 0xf5, G: 0xf0, B: 0xe8, A: 0xff}
	lightForeground  = color.NRGBA{R: 0x2c, G: 0x2c, B: 0x2c, A: 0xff}
	lightButton      = color.NRGBA{R: 0xe8, G: 0xe0, B: 0xd0, A: 0xff}
	lightAccent      = color.NRGBA{R: 0x6d, G: 0x28, B: 0xd9, A: 0xff}
	lightInputBg     = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	lightPlaceholder = color.NRGBA{R: 0x9c, G: 0x9c, B: 0x9c, A: 0xff}
)

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

// ApplyTheme rebuilds and applies the theme to the given Fyne app.
// Call this whenever DarkMode or AppSize changes.
func ApplyTheme(app fyne.App) {
	app.Settings().SetTheme(NewTheme(config.Options.AppSize))
}

// Color returns the color for the specified theme color name and variant.
func (t *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	dark := config.Options.DarkMode

	switch name {
	case theme.ColorNameBackground:
		if dark {
			return darkBackground
		}
		return lightBackground

	case theme.ColorNameForeground:
		if dark {
			return darkForeground
		}
		return lightForeground

	case theme.ColorNameButton:
		if dark {
			return darkButton
		}
		return lightButton

	case theme.ColorNamePrimary:
		if dark {
			return darkAccent
		}
		return lightAccent

	case theme.ColorNameInputBackground:
		if dark {
			return darkInputBg
		}
		return lightInputBg

	case theme.ColorNamePlaceHolder:
		if dark {
			return darkPlaceholder
		}
		return lightPlaceholder

	case theme.ColorNameHover:
		if dark {
			return color.NRGBA{R: 0xcb, G: 0xa6, B: 0xf7, A: 0x30}
		}
		return color.NRGBA{R: 0x6d, G: 0x28, B: 0xd9, A: 0x20}

	case theme.ColorNameDisabled:
		if dark {
			return color.NRGBA{R: 0x58, G: 0x5b, B: 0x70, A: 0xff}
		}
		return color.NRGBA{R: 0xb0, G: 0xb0, B: 0xb0, A: 0xff}
	}

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
		return float32(t.baseSize)
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

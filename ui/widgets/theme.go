package widgets

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed SimpleCLM-Medium.ttf
var SimpleCLMMedium []byte

type Theme struct{}

func (t *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (t *Theme) Font(s fyne.TextStyle) fyne.Resource {
	return fyne.NewStaticResource("SimpleCLM-Medium.ttf", SimpleCLMMedium)
}

func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	// if name == fyne.ThemeSizeName("innerPadding") {
	// 	return 5
	// }
	// if name == fyne.ThemeSizeName("iconInline") {
	// 	return 20
	// }
	// if name == fyne.ThemeSizeName("text") {
	// 	return 16
	// }
	return theme.DefaultTheme().Size(name)
}

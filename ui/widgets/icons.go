package widgets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed load.svg
var loadIcon []byte

// LoadIcon load icon
var LoadIcon = fyne.NewStaticResource("load.svg", loadIcon)

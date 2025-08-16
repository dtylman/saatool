package widgets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed settings.svg
var iconSettings []byte
var IconSettings = fyne.NewStaticResource("settings.svg", iconSettings)

//go:embed project.svg
var iconProject []byte
var IconProject = fyne.NewStaticResource("project.svg", iconProject)

//go:embed logo.svg
var iconLogo []byte
var IconLogo = fyne.NewStaticResource("logo.svg", iconLogo)

//go:embed log.svg
var iconLog []byte
var IconLog = fyne.NewStaticResource("log.svg", iconLog)

//go:embed open.svg
var iconOpen []byte
var IconOpen = fyne.NewStaticResource("open.svg", iconOpen)

//go:embed translate.svg
var iconTranslate []byte
var IconTranslate = fyne.NewStaticResource("translate.svg", iconTranslate)

//go:embed next.svg
var iconNext []byte
var IconNext = fyne.NewStaticResource("next.svg", iconNext)

//go:embed prev.svg
var iconPrev []byte
var IconPrev = fyne.NewStaticResource("prev.svg", iconPrev)

package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed icons/idle.svg
var idleIconContent []byte

//go:embed icons/running.svg
var runningIconContent []byte

var (
	// AppIcon is also installed as the desktop menu icon by the Makefile.
	AppIcon     = fyne.NewStaticResource("meanwhile.svg", idleIconContent)
	runningIcon = fyne.NewStaticResource("meanwhile-running.svg", runningIconContent)
)

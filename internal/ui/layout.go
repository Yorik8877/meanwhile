package ui

import "fyne.io/fyne/v2"

const (
	formWidth = 420
	// settingsPanelMargin keeps the settings panel off the window's edge.
	settingsPanelMargin = 8
	// settingsPanelPadding is extra space above, below and to the right of
	// the settings inside the panel.
	settingsPanelPadding float32 = 8
)

var (
	windowSize         = fyne.NewSize(520, 640)
	intervalEditorSize = fyne.NewSize(480, 420)
	startupErrorSize   = fyne.NewSize(480, 180)
)

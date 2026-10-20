package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// startIconScale makes the start/pause button the most prominent control
// on a card.
const startIconScale = 1.5

// startButton starts and pauses a task. A button draws its icon at the
// theme's inline icon size, so the larger icon lies on top of an empty
// button that keeps the hover, focus and tap behaviour.
type startButton struct {
	widget.BaseWidget
	OnTapped func()
	tapArea  *widget.Button
	backdrop *cardButton
	icon     *widget.Icon
}

func newStartButton() *startButton {
	button := &startButton{icon: widget.NewIcon(theme.MediaPlayIcon())}
	button.tapArea = widget.NewButton("", func() {
		if button.OnTapped != nil {
			button.OnTapped()
		}
	})
	button.backdrop = newCardButton(button.tapArea)
	button.ExtendBaseWidget(button)
	return button
}

// SetRunning shows the pause icon on the primary colour while the task runs
// and the play icon otherwise.
func (button *startButton) SetRunning(running bool) {
	if running {
		button.tapArea.Importance = widget.HighImportance
		button.icon.Resource = theme.NewColoredResource(theme.MediaPauseIcon(), theme.ColorNameForegroundOnPrimary)
	} else {
		button.tapArea.Importance = widget.LowImportance
		button.icon.Resource = theme.MediaPlayIcon()
	}
	button.tapArea.Refresh()
	button.icon.Refresh()
}

func (button *startButton) CreateRenderer() fyne.WidgetRenderer {
	iconSize := theme.IconInlineSize() * startIconScale
	padding := theme.InnerPadding()
	iconLayer := container.New(layout.NewCustomPaddedLayout(padding, padding, padding, padding),
		container.NewCenter(container.NewGridWrap(fyne.NewSquareSize(iconSize), button.icon)))
	return widget.NewSimpleRenderer(container.NewStack(button.backdrop, iconLayer))
}

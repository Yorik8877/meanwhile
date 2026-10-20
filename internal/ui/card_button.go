package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// cardButton puts a standard button on the window background, so that it
// stands out from the lighter card. The button keeps its hover, focus and
// tap behaviour; it is made low importance, which leaves its own background
// transparent until it is hovered or focused.
type cardButton struct {
	widget.BaseWidget
	button *widget.Button
}

func newCardButton(button *widget.Button) *cardButton {
	button.Importance = widget.LowImportance
	newButton := &cardButton{button: button}
	newButton.ExtendBaseWidget(newButton)
	return newButton
}

func (button *cardButton) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(cardButtonColor())
	background.CornerRadius = theme.Size(theme.SizeNameButtonRadius)
	return &cardButtonRenderer{
		WidgetRenderer: widget.NewSimpleRenderer(container.NewStack(background, button.button)),
		background:     background,
	}
}

// cardButtonRenderer repaints the background on refresh, so that the
// button follows a switch between the light and the dark theme.
type cardButtonRenderer struct {
	fyne.WidgetRenderer
	background *canvas.Rectangle
}

func (renderer *cardButtonRenderer) Refresh() {
	renderer.background.FillColor = cardButtonColor()
	renderer.background.CornerRadius = theme.Size(theme.SizeNameButtonRadius)
	renderer.background.Refresh()
	renderer.WidgetRenderer.Refresh()
}

func cardButtonColor() color.Color {
	return theme.Color(theme.ColorNameBackground)
}

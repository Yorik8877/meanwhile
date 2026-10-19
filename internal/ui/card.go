package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// card shows content on a rounded background that sets one task apart
// from the next, with the contrast overrides applied to the content.
type card struct {
	widget.BaseWidget
	content fyne.CanvasObject
}

func newCard(content fyne.CanvasObject) *card {
	newCard := &card{content: content}
	newCard.ExtendBaseWidget(newCard)
	return newCard
}

func (card *card) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(cardColor())
	background.CornerRadius = theme.InputRadiusSize()
	padded := container.NewPadded(container.NewThemeOverride(card.content, contrastTheme{}))
	return &cardRenderer{
		WidgetRenderer: widget.NewSimpleRenderer(container.NewStack(background, padded)),
		background:     background,
	}
}

// cardRenderer repaints the background on refresh, so that the card follows
// a switch between the light and the dark theme.
type cardRenderer struct {
	fyne.WidgetRenderer
	background *canvas.Rectangle
}

func (renderer *cardRenderer) Refresh() {
	renderer.background.FillColor = cardColor()
	renderer.background.CornerRadius = theme.InputRadiusSize()
	renderer.background.Refresh()
	renderer.WidgetRenderer.Refresh()
}

// cardColor is the background of text fields: slightly lighter than the
// window in the dark theme and slightly darker in the light one.
func cardColor() color.Color {
	return theme.Color(theme.ColorNameInputBackground)
}

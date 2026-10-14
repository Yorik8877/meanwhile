package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// cardTheme paints buttons on a task card with the window background, so
// that they stand out from the lighter card instead of merging with it.
// A non-zero iconScale enlarges button icons.
type cardTheme struct {
	iconScale float32
}

func (overrides cardTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameButton:
		name = theme.ColorNameBackground
	case theme.ColorNameDisabled:
		// Low importance labels use the disabled colour, which is too faint
		// for the note line; nothing on a card is ever disabled.
		name = theme.ColorNamePlaceHolder
	}
	return currentTheme().Color(name, variant)
}

func (overrides cardTheme) Font(style fyne.TextStyle) fyne.Resource {
	return currentTheme().Font(style)
}

func (overrides cardTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return currentTheme().Icon(name)
}

func (overrides cardTheme) Size(name fyne.ThemeSizeName) float32 {
	size := currentTheme().Size(name)
	if name == theme.SizeNameInlineIcon && overrides.iconScale > 0 {
		return size * overrides.iconScale
	}
	return size
}

// currentTheme is looked up on every call, so that the overrides follow a
// switch between the light and the dark theme.
func currentTheme() fyne.Theme {
	return fyne.CurrentApp().Settings().Theme()
}

package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// contrastTheme raises contrast where the standard theme is too soft:
// buttons get the window background, so that they stand out from a lighter
// task card, and low importance text gets the readable placeholder colour.
// A non-zero iconScale enlarges button icons.
type contrastTheme struct {
	iconScale float32
}

func (overrides contrastTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameButton:
		name = theme.ColorNameBackground
	case theme.ColorNameDisabled:
		// Low importance labels use the disabled colour, which is too faint
		// to read. Wrap only content that has no disabled widgets.
		name = theme.ColorNamePlaceHolder
	}
	return currentTheme().Color(name, variant)
}

func (overrides contrastTheme) Font(style fyne.TextStyle) fyne.Resource {
	return currentTheme().Font(style)
}

func (overrides contrastTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return currentTheme().Icon(name)
}

func (overrides contrastTheme) Size(name fyne.ThemeSizeName) float32 {
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

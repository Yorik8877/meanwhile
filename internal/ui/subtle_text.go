package ui

import (
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// newSubtleText shows secondary text, such as a note or a hint, in the
// placeholder colour: quieter than ordinary text, yet easy to read, unlike
// a low importance label, which takes the faint disabled colour.
func newSubtleText(text string) *widget.RichText {
	style := widget.RichTextStyleInline
	style.ColorName = theme.ColorNamePlaceHolder
	return widget.NewRichText(&widget.TextSegment{Text: text, Style: style})
}

// setSubtleText replaces the text shown by newSubtleText.
func setSubtleText(subtleText *widget.RichText, text string) {
	subtleText.Segments[0].(*widget.TextSegment).Text = text
	subtleText.Refresh()
}

package ui

import (
	"errors"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// reportError logs the error and shows it to the user in the window.
func reportError(window fyne.Window, err error) {
	if err == nil {
		return
	}
	slog.Error("action failed", "error", err)
	dialog.ShowError(errors.New(errorText(err)), window)
}

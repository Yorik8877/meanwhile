package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowStartupError shows why the program cannot start and blocks until the
// user closes the window.
func ShowStartupError(application fyne.App, databasePath string, err error) {
	window := application.NewWindow(message{"app.title", "Stopwatch"}.String())
	text := widget.NewLabel(message{"startup.database_error", "Cannot open the database {{.Path}}: {{.Error}}"}.With(map[string]any{
		"Path":  databasePath,
		"Error": err.Error(),
	}))
	text.Wrapping = fyne.TextWrapWord
	closeButton := widget.NewButton(message{"form.close", "Close"}.String(), application.Quit)

	window.SetContent(container.NewBorder(nil, closeButton, nil, nil, text))
	window.Resize(startupErrorSize)
	window.SetMaster()
	window.ShowAndRun()
}

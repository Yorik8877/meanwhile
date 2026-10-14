package ui

import (
	"log/slog"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Settings are the stored preferences shown in the settings panel.
// Autostart is nil when it must not be changed, as in a development copy.
type Settings struct {
	Language  LanguageSetting
	Autostart AutostartSetting
}

type AutostartSetting interface {
	Enabled() (bool, error)
	SetEnabled(enabled bool) error
}

// showSettings opens a small panel under the settings button, kept away
// from the window's right edge.
func (mainWindow *MainWindow) showSettings(anchor fyne.CanvasObject) {
	var panel *widget.PopUp
	languageRow, err := mainWindow.languageRow(func() { panel.Hide() })
	if err != nil {
		reportError(mainWindow.window, err)
		return
	}
	content := container.NewVBox(languageRow)
	if mainWindow.settings.Autostart != nil {
		content.Add(mainWindow.autostartCheck())
	}

	windowCanvas := mainWindow.window.Canvas()
	panel = widget.NewPopUp(content, windowCanvas)
	anchorPosition := fyne.CurrentApp().Driver().AbsolutePositionForObject(anchor)
	panelLeft := windowCanvas.Size().Width - panel.MinSize().Width - settingsPanelMargin
	panel.ShowAtPosition(fyne.NewPos(panelLeft, anchorPosition.Y+anchor.Size().Height))
}

// languageRow is "Language: [selector]". Choosing a language rebuilds the
// window, so the panel is closed first.
func (mainWindow *MainWindow) languageRow(closePanel func()) (fyne.CanvasObject, error) {
	current, err := mainWindow.settings.Language.Language()
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(languageOptions))
	for index, option := range languageOptions {
		labels[index] = option.label()
	}
	selector := widget.NewSelect(labels, nil)
	selector.SetSelected(languageOptionFor(current).label())
	// Assigned after SetSelected, so that showing the current value does not
	// count as a change.
	selector.OnChanged = func(label string) {
		closePanel()
		mainWindow.switchLanguage(languageOptions[slices.Index(labels, label)].code)
	}

	label := widget.NewLabel(message{"settings.language", "Language:"}.String())
	return container.New(layout.NewFormLayout(), label, selector), nil
}

// autostartCheck reverts itself when the change cannot be saved, so that it
// always shows the real state.
func (mainWindow *MainWindow) autostartCheck() *widget.Check {
	autostart := mainWindow.settings.Autostart
	check := widget.NewCheck(message{"settings.autostart", "Start at login"}.String(), nil)
	enabled, err := autostart.Enabled()
	if err != nil {
		slog.Error("read autostart state", "error", err)
	}
	check.SetChecked(enabled)

	var onChanged func(bool)
	onChanged = func(enabled bool) {
		err := autostart.SetEnabled(enabled)
		if err == nil {
			return
		}
		reportError(mainWindow.window, err)
		check.OnChanged = nil
		check.SetChecked(!enabled)
		check.OnChanged = onChanged
	}
	check.OnChanged = onChanged
	return check
}

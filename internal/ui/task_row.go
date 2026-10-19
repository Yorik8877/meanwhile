package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

// toggleIconScale makes the start/pause button the most prominent control
// on a card.
const toggleIconScale = 1.5

// taskRow is one task card in the task list. The list reuses rows, so show
// binds a row to a task every time it is redrawn.
type taskRow struct {
	widget.BaseWidget
	toggle    *widget.Button
	name      *widget.Label
	note      *widget.Label
	times     *widget.Label
	intervals *widget.Button
	edit      *widget.Button
	complete  *widget.Button
	remove    *widget.Button
}

// taskActions handle the buttons of one card.
type taskActions struct {
	toggle    func()
	intervals func()
	edit      func()
	complete  func()
	remove    func()
}

func newTaskRow() *taskRow {
	row := &taskRow{
		toggle: widget.NewButtonWithIcon("", theme.MediaPlayIcon(), nil),
		name:   widget.NewLabel(""),
		note:   widget.NewLabel(""),
		// Two lines from the start: the list sizes every row by the template row.
		times:     widget.NewLabel("\n"),
		intervals: widget.NewButton(message{"tasks.intervals", "Intervals"}.String(), nil),
		edit:      widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil),
		complete:  widget.NewButtonWithIcon("", theme.ConfirmIcon(), nil),
		remove:    widget.NewButtonWithIcon("", theme.NewErrorThemedResource(theme.DeleteIcon()), nil),
	}
	row.name.Truncation = fyne.TextTruncateEllipsis
	row.name.SizeName = theme.SizeNameSubHeadingText
	row.note.Truncation = fyne.TextTruncateEllipsis
	row.note.Importance = widget.LowImportance
	row.times.Alignment = fyne.TextAlignTrailing
	row.ExtendBaseWidget(row)
	return row
}

func (row *taskRow) CreateRenderer() fyne.WidgetRenderer {
	details := container.NewVBox(row.name, row.note)
	buttons := container.NewHBox(layout.NewSpacer(), row.intervals, row.edit, row.complete, row.remove)
	times := container.NewVBox(buttons, row.times)
	toggle := container.NewThemeOverride(row.toggle, contrastTheme{iconScale: toggleIconScale})
	content := container.NewBorder(nil, nil, container.NewVBox(toggle), times, details)
	return widget.NewSimpleRenderer(newCard(content))
}

func (row *taskRow) show(summary tracking.TaskSummary, actions taskActions) {
	row.name.SetText(summary.Task.Name)
	row.name.TextStyle.Bold = summary.Running
	row.note.SetText(firstLine(summary.Task.Note))
	// One two-line label keeps the lines closer than two labels would.
	row.times.SetText(message{"tasks.today", "Today {{.Duration}}"}.With(map[string]any{
		"Duration": formatDuration(summary.Today),
	}) + "\n" + message{"tasks.total", "Total {{.Duration}}"}.With(map[string]any{
		"Duration": formatDuration(summary.Total),
	}))

	if summary.Running {
		row.toggle.SetIcon(theme.MediaPauseIcon())
		row.toggle.Importance = widget.HighImportance
	} else {
		row.toggle.SetIcon(theme.MediaPlayIcon())
		row.toggle.Importance = widget.MediumImportance
	}
	row.toggle.OnTapped = actions.toggle
	row.intervals.OnTapped = actions.intervals
	row.edit.OnTapped = actions.edit
	row.complete.OnTapped = actions.complete
	row.remove.OnTapped = actions.remove
	row.Refresh()
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return line
}

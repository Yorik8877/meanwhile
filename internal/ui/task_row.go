package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"stopwatch/internal/tracking"
)

// taskRow is one line of the task list. The list reuses rows, so show binds
// a row to a task every time it is redrawn.
type taskRow struct {
	widget.BaseWidget
	toggle *widget.Button
	name   *widget.Label
	note   *widget.Label
	times  *widget.Label
	more   *widget.Button
}

func newTaskRow() *taskRow {
	row := &taskRow{
		toggle: widget.NewButtonWithIcon("", theme.MediaPlayIcon(), nil),
		name:   widget.NewLabel(""),
		note:   widget.NewLabel(""),
		times:  widget.NewLabel(""),
		more:   widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), nil),
	}
	row.name.Truncation = fyne.TextTruncateEllipsis
	row.note.Truncation = fyne.TextTruncateEllipsis
	row.note.Importance = widget.LowImportance
	row.ExtendBaseWidget(row)
	return row
}

func (row *taskRow) CreateRenderer() fyne.WidgetRenderer {
	details := container.NewVBox(row.name, row.note, row.times)
	return widget.NewSimpleRenderer(container.NewBorder(nil, nil, row.toggle, row.more, details))
}

func (row *taskRow) show(summary tracking.TaskSummary, onToggle func(), onMore func(anchor fyne.CanvasObject)) {
	row.name.SetText(summary.Task.Name)
	row.name.TextStyle.Bold = summary.Running
	row.note.SetText(firstLine(summary.Task.Note))
	row.times.SetText(message{"tasks.times", "Today {{.Today}} · Total {{.Total}}"}.With(map[string]any{
		"Today": formatDuration(summary.Today),
		"Total": formatDuration(summary.Total),
	}))

	if summary.Running {
		row.toggle.SetIcon(theme.MediaPauseIcon())
		row.toggle.Importance = widget.HighImportance
	} else {
		row.toggle.SetIcon(theme.MediaPlayIcon())
		row.toggle.Importance = widget.MediumImportance
	}
	row.toggle.OnTapped = onToggle
	row.more.OnTapped = func() { onMore(row.more) }
	row.Refresh()
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return line
}

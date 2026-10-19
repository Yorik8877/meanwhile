package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

// completedRow is the card of a completed task. It has no start button:
// a completed task is reopened first.
type completedRow struct {
	widget.BaseWidget
	name      *widget.Label
	note      *widget.Label
	details   *widget.Label
	reopen    *widget.Button
	intervals *widget.Button
	remove    *widget.Button
}

// completionMomentLayout shows the completion moment as "2026-10-19 13:45".
const completionMomentLayout = "2006-01-02 15:04"

type completedActions struct {
	reopen    func()
	intervals func()
	remove    func()
}

func newCompletedRow() *completedRow {
	row := &completedRow{
		name: widget.NewLabel(""),
		note: widget.NewLabel(""),
		// Two lines from the start: the list sizes every row by the template row.
		details:   widget.NewLabel("\n"),
		reopen:    widget.NewButton(message{"completed.reopen", "Reopen"}.String(), nil),
		intervals: widget.NewButton(message{"tasks.intervals", "Intervals"}.String(), nil),
		remove:    widget.NewButtonWithIcon("", theme.NewErrorThemedResource(theme.DeleteIcon()), nil),
	}
	row.name.Truncation = fyne.TextTruncateEllipsis
	row.name.SizeName = theme.SizeNameSubHeadingText
	row.note.Truncation = fyne.TextTruncateEllipsis
	row.note.Importance = widget.LowImportance
	row.details.Alignment = fyne.TextAlignTrailing
	row.ExtendBaseWidget(row)
	return row
}

func (row *completedRow) CreateRenderer() fyne.WidgetRenderer {
	buttons := container.NewHBox(layout.NewSpacer(), row.reopen, row.intervals, row.remove)
	content := container.NewBorder(nil, nil, nil, container.NewVBox(buttons, row.details), container.NewVBox(row.name, row.note))
	return widget.NewSimpleRenderer(newCard(content))
}

func (row *completedRow) show(summary tracking.TaskSummary, actions completedActions) {
	row.name.SetText(summary.Task.Name)
	row.note.SetText(firstLine(summary.Task.Note))
	row.details.SetText(message{"completed.completed_on", "Completed {{.Moment}}"}.With(map[string]any{
		"Moment": summary.Task.CompletedAt.In(time.Local).Format(completionMomentLayout),
	}) + "\n" + message{"tasks.total", "Total {{.Duration}}"}.With(map[string]any{
		"Duration": formatDuration(summary.Total),
	}))
	row.reopen.OnTapped = actions.reopen
	row.intervals.OnTapped = actions.intervals
	row.remove.OnTapped = actions.remove
	row.Refresh()
}

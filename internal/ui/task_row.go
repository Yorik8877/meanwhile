package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"stopwatch/internal/tracking"
)

// toggleIconScale makes the start/pause button the most prominent control
// on a card.
const toggleIconScale = 1.5

// taskRow is one task card in the task list. The list reuses rows, so show
// binds a row to a task every time it is redrawn.
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
		// Two lines from the start: the list sizes every row by the template row.
		times: widget.NewLabel("\n"),
		more:  widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), nil),
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
	background := canvas.NewRectangle(cardColor())
	background.CornerRadius = theme.InputRadiusSize()

	details := container.NewVBox(row.name, row.note)
	times := container.NewVBox(container.NewHBox(layout.NewSpacer(), row.more), row.times)
	toggle := container.NewThemeOverride(row.toggle, cardTheme{iconScale: toggleIconScale})
	content := container.NewBorder(nil, nil, container.NewVBox(toggle), times, details)
	card := container.NewStack(background, container.NewPadded(container.NewThemeOverride(content, cardTheme{})))
	return &taskRowRenderer{WidgetRenderer: widget.NewSimpleRenderer(card), background: background}
}

func (row *taskRow) show(summary tracking.TaskSummary, onToggle func(), onMore func(anchor fyne.CanvasObject)) {
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
	row.toggle.OnTapped = onToggle
	row.more.OnTapped = func() { onMore(row.more) }
	row.Refresh()
}

// taskRowRenderer repaints the card background on refresh, so that the
// card follows a switch between the light and the dark theme.
type taskRowRenderer struct {
	fyne.WidgetRenderer
	background *canvas.Rectangle
}

func (renderer *taskRowRenderer) Refresh() {
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

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return line
}

package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"stopwatch/internal/tracking"
)

type intervalEditor struct {
	tracker *tracking.Tracker
	window  fyne.Window
	task    tracking.Task
	rows    *fyne.Container
}

func showIntervalEditor(tracker *tracking.Tracker, window fyne.Window, task tracking.Task) {
	editor := &intervalEditor{tracker: tracker, window: window, task: task, rows: container.NewVBox()}
	addButton := widget.NewButtonWithIcon(message{"intervals.add", "Add interval"}.String(), theme.ContentAddIcon(), func() {
		editor.showForm(tracking.Interval{TaskID: task.ID})
	})
	content := container.NewBorder(nil, addButton, nil, nil, container.NewVScroll(editor.rows))

	editorDialog := dialog.NewCustom(
		message{"intervals.title", "Intervals: {{.Name}}"}.With(map[string]any{"Name": task.Name}),
		message{"form.close", "Close"}.String(),
		content,
		window,
	)
	editorDialog.Resize(intervalEditorSize)
	editor.rebuild()
	editorDialog.Show()
}

func (editor *intervalEditor) rebuild() {
	editor.rows.RemoveAll()
	intervals := editor.tracker.Intervals(editor.task.ID)
	if len(intervals) == 0 {
		editor.rows.Add(widget.NewLabel(message{"intervals.empty", "No intervals yet."}.String()))
		return
	}
	now := time.Now()
	for _, interval := range intervals {
		editor.rows.Add(editor.row(interval, now))
	}
}

func (editor *intervalEditor) row(interval tracking.Interval, now time.Time) fyne.CanvasObject {
	description := widget.NewLabel(describeInterval(interval, now))
	editButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() { editor.showForm(interval) })
	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() { editor.confirmDelete(interval) })
	return container.NewBorder(nil, nil, nil, container.NewHBox(editButton, deleteButton), description)
}

func (editor *intervalEditor) showForm(interval tracking.Interval) {
	form := newIntervalForm(interval, time.Now())
	title := message{"interval_form.edit_title", "Edit interval"}.String()
	if interval.ID == 0 {
		title = message{"interval_form.add_title", "New interval"}.String()
	}
	formDialog := dialog.NewCustomWithoutButtons(title, form.content, editor.window)

	saveButton := widget.NewButton(message{"form.save", "Save"}.String(), func() {
		if err := editor.save(interval, form); err != nil {
			form.showError(err)
			return
		}
		formDialog.Hide()
		editor.rebuild()
	})
	saveButton.Importance = widget.HighImportance
	cancelButton := widget.NewButton(message{"form.cancel", "Cancel"}.String(), formDialog.Hide)
	formDialog.SetButtons([]fyne.CanvasObject{cancelButton, saveButton})
	formDialog.Resize(fyne.NewSize(formWidth, formDialog.MinSize().Height))
	formDialog.Show()
}

func (editor *intervalEditor) save(original tracking.Interval, form *intervalForm) error {
	start, end, err := form.moments(original)
	if err != nil {
		return err
	}
	if original.ID == 0 {
		return editor.tracker.AddInterval(editor.task.ID, start, end)
	}
	return editor.tracker.UpdateInterval(original.ID, start, end)
}

func (editor *intervalEditor) confirmDelete(interval tracking.Interval) {
	dialog.ShowConfirm(
		message{"intervals.delete_title", "Delete interval"}.String(),
		message{"intervals.delete_question", "Delete this interval?"}.String(),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			reportError(editor.window, editor.tracker.DeleteInterval(interval.ID))
			editor.rebuild()
		},
		editor.window,
	)
}

// describeInterval gives "Oct 13, 09:00 – 10:30 (1:30:00)", or
// "Oct 13, 09:00 – running" for a running interval.
func describeInterval(interval tracking.Interval, now time.Time) string {
	start := interval.StartedAt.In(time.Local)
	date := formatShortDate(tracking.DateOf(start))
	if interval.IsRunning() {
		return fmt.Sprintf("%s, %s – %s", date, formatClock(start), message{"intervals.running", "running"}.String())
	}
	end := interval.StoppedAt.In(time.Local)
	return fmt.Sprintf("%s, %s – %s (%s)", date, formatClock(start), formatClock(end), formatDuration(interval.DurationAt(now)))
}

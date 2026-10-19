package ui

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

func showTaskForm(tracker *tracking.Tracker, window fyne.Window, task tracking.Task) {
	name := widget.NewEntry()
	name.SetText(task.Name)
	name.Validator = func(text string) error {
		if _, err := tracking.ValidateTaskName(text); err != nil {
			return errors.New(errorText(err))
		}
		return nil
	}
	note := widget.NewMultiLineEntry()
	note.SetText(task.Note)
	note.SetMinRowsVisible(4)

	items := []*widget.FormItem{
		widget.NewFormItem(message{"task_form.name", "Name"}.String(), name),
		widget.NewFormItem(message{"task_form.note", "Note"}.String(), note),
	}
	form := dialog.NewForm(
		message{"task_form.title", "Edit task"}.String(),
		message{"form.save", "Save"}.String(),
		message{"form.cancel", "Cancel"}.String(),
		items,
		func(confirmed bool) {
			if !confirmed {
				return
			}
			reportError(window, tracker.UpdateTask(task.ID, name.Text, note.Text))
		},
		window,
	)
	form.Resize(fyne.NewSize(formWidth, form.MinSize().Height))
	form.Show()
}

// confirmTaskDeletion asks first, because the task's recorded time goes
// together with it.
func confirmTaskDeletion(tracker *tracking.Tracker, window fyne.Window, task tracking.Task) {
	dialog.ShowConfirm(
		message{"tasks.delete_title", "Delete task"}.String(),
		message{"tasks.delete_question", "Delete “{{.Name}}” together with all its recorded time?"}.With(map[string]any{"Name": task.Name}),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			reportError(window, tracker.DeleteTask(task.ID))
		},
		window,
	)
}

package ui

import (
	"errors"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"stopwatch/internal/tracking"
)

type tasksTab struct {
	tracker   *tracking.Tracker
	window    fyne.Window
	summaries []tracking.TaskSummary
	newTask   *widget.Entry
	list      *widget.List
	pauseAll  *widget.Button
	content   fyne.CanvasObject
}

func newTasksTab(tracker *tracking.Tracker, window fyne.Window) *tasksTab {
	tab := &tasksTab{tracker: tracker, window: window}

	tab.newTask = widget.NewEntry()
	tab.newTask.SetPlaceHolder(message{"tasks.new_placeholder", "New task"}.String())
	tab.newTask.OnSubmitted = func(string) { tab.addTask() }
	addButton := widget.NewButton(message{"tasks.add", "Add"}.String(), tab.addTask)

	tab.list = widget.NewList(
		func() int { return len(tab.summaries) },
		func() fyne.CanvasObject { return newTaskRow() },
		tab.showRow,
	)
	tab.pauseAll = widget.NewButton(message{"tasks.pause_all", "Pause all"}.String(), func() {
		reportError(tab.window, tab.tracker.PauseAll(time.Now()))
	})

	top := container.NewBorder(nil, nil, nil, addButton, tab.newTask)
	tab.content = container.NewBorder(top, tab.pauseAll, nil, nil, tab.list)
	tab.refresh()
	return tab
}

func (tab *tasksTab) refresh() {
	tab.summaries = tab.tracker.TaskSummaries()
	tab.list.Refresh()

	anyRunning := false
	for _, summary := range tab.summaries {
		anyRunning = anyRunning || summary.Running
	}
	if anyRunning {
		tab.pauseAll.Enable()
	} else {
		tab.pauseAll.Disable()
	}
}

func (tab *tasksTab) addTask() {
	if _, err := tab.tracker.CreateTask(tab.newTask.Text, ""); err != nil {
		reportError(tab.window, err)
		return
	}
	tab.newTask.SetText("")
}

func (tab *tasksTab) showRow(index widget.ListItemID, object fyne.CanvasObject) {
	if index >= len(tab.summaries) {
		return
	}
	task := tab.summaries[index].Task
	object.(*taskRow).show(tab.summaries[index], taskActions{
		toggle:    func() { reportError(tab.window, tab.tracker.Toggle(task.ID)) },
		intervals: func() { showIntervalEditor(tab.tracker, tab.window, task) },
		edit:      func() { tab.showTaskForm(task) },
		remove:    func() { tab.confirmDelete(task) },
	})
}

func (tab *tasksTab) showTaskForm(task tracking.Task) {
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
			reportError(tab.window, tab.tracker.UpdateTask(task.ID, name.Text, note.Text))
		},
		tab.window,
	)
	form.Resize(fyne.NewSize(formWidth, form.MinSize().Height))
	form.Show()
}

func (tab *tasksTab) confirmDelete(task tracking.Task) {
	dialog.ShowConfirm(
		message{"tasks.delete_title", "Delete task"}.String(),
		message{"tasks.delete_question", "Delete “{{.Name}}” together with all its recorded time?"}.With(map[string]any{"Name": task.Name}),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			reportError(tab.window, tab.tracker.DeleteTask(task.ID))
		},
		tab.window,
	)
}

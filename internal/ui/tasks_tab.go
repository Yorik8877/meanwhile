package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/yorik8877/meanwhile/internal/tracking"
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
	tab.summaries = tab.tracker.ActiveTaskSummaries()
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
		edit:      func() { showTaskForm(tab.tracker, tab.window, task) },
		complete:  func() { reportError(tab.window, tab.tracker.Complete(task.ID)) },
		remove:    func() { confirmTaskDeletion(tab.tracker, tab.window, task) },
	})
}

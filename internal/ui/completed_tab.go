package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

// completedTab keeps finished tasks out of the way but within reach, so
// that one can be reopened when work on it comes back.
type completedTab struct {
	tracker   *tracking.Tracker
	window    fyne.Window
	summaries []tracking.TaskSummary
	list      *widget.List
	empty     *widget.Label
	content   fyne.CanvasObject
}

func newCompletedTab(tracker *tracking.Tracker, window fyne.Window) *completedTab {
	tab := &completedTab{tracker: tracker, window: window}
	tab.list = widget.NewList(
		func() int { return len(tab.summaries) },
		func() fyne.CanvasObject { return newCompletedRow() },
		tab.showRow,
	)
	tab.empty = widget.NewLabel(message{"completed.empty", "No completed tasks."}.String())
	tab.empty.Importance = widget.LowImportance
	readableEmpty := container.NewThemeOverride(container.NewCenter(tab.empty), contrastTheme{})
	tab.content = container.NewStack(tab.list, readableEmpty)
	tab.refresh()
	return tab
}

func (tab *completedTab) refresh() {
	tab.summaries = tab.tracker.CompletedTaskSummaries()
	tab.list.Refresh()
	if len(tab.summaries) == 0 {
		tab.empty.Show()
	} else {
		tab.empty.Hide()
	}
}

func (tab *completedTab) reopen(task tracking.Task) {
	reportError(tab.window, tab.tracker.Reopen(task.ID))
}

func (tab *completedTab) showRow(index widget.ListItemID, object fyne.CanvasObject) {
	if index >= len(tab.summaries) {
		return
	}
	task := tab.summaries[index].Task
	object.(*completedRow).show(tab.summaries[index], completedActions{
		reopen:    func() { tab.reopen(task) },
		intervals: func() { showIntervalEditor(tab.tracker, tab.window, task) },
		remove:    func() { confirmTaskDeletion(tab.tracker, tab.window, task) },
	})
}

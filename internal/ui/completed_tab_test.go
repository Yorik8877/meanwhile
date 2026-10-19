package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestCompletedTabListsCompletedTasksAndReopensThem(t *testing.T) {
	test.NewTempApp(t)
	tracker := newTestTracker(t)
	window := test.NewTempWindow(t, nil)
	tab := newCompletedTab(tracker, window)
	window.SetContent(tab.content)
	if !tab.empty.Visible() {
		t.Error("empty hint is hidden although nothing is completed")
	}
	id, err := tracker.CreateTask("Task", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := tracker.Complete(id); err != nil {
		t.Fatal(err)
	}

	tab.refresh()
	if len(tab.summaries) != 1 || tab.empty.Visible() {
		t.Fatalf("completed tab shows %d tasks, empty hint visible = %v; want 1 task and no hint",
			len(tab.summaries), tab.empty.Visible())
	}
	tab.reopen(tab.summaries[0].Task)
	tab.refresh()

	if len(tab.summaries) != 0 || len(tracker.ActiveTaskSummaries()) != 1 {
		t.Errorf("after reopening: %d completed, %d active; want 0 and 1",
			len(tab.summaries), len(tracker.ActiveTaskSummaries()))
	}
}

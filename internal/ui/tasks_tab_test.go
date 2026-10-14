package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"stopwatch/internal/tracking"
	"stopwatch/internal/tracking/memory"
)

func newTestTracker(t *testing.T) *tracking.Tracker {
	t.Helper()
	tracker := tracking.NewTracker(memory.NewRepository(), tracking.SystemClock{}, time.UTC)
	if err := tracker.Load(); err != nil {
		t.Fatal(err)
	}
	return tracker
}

func TestAddingTaskFromEntry(t *testing.T) {
	test.NewTempApp(t)
	tracker := newTestTracker(t)
	window := test.NewTempWindow(t, nil)
	tab := newTasksTab(tracker, window)
	window.SetContent(tab.content)

	test.Type(tab.newTask, "Write tests")
	tab.newTask.OnSubmitted(tab.newTask.Text)

	summaries := tracker.TaskSummaries()
	if len(summaries) != 1 || summaries[0].Task.Name != "Write tests" {
		t.Fatalf("summaries = %+v, want one task named Write tests", summaries)
	}
	if tab.newTask.Text != "" {
		t.Errorf("entry = %q, want it cleared", tab.newTask.Text)
	}
}

func TestPauseAllButtonFollowsRunningState(t *testing.T) {
	test.NewTempApp(t)
	tracker := newTestTracker(t)
	window := test.NewTempWindow(t, nil)
	tab := newTasksTab(tracker, window)
	window.SetContent(tab.content)
	id, err := tracker.CreateTask("Task", "")
	if err != nil {
		t.Fatal(err)
	}

	if tab.refresh(); !tab.pauseAll.Disabled() {
		t.Error("pause all is enabled while nothing runs")
	}
	if err := tracker.Start(id); err != nil {
		t.Fatal(err)
	}
	if tab.refresh(); tab.pauseAll.Disabled() {
		t.Fatal("pause all is disabled while a task runs")
	}

	test.Tap(tab.pauseAll)

	if tracker.AnyRunning() {
		t.Error("task still runs after tapping pause all")
	}
}

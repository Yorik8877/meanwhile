package ui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/yorik8877/meanwhile/internal/tracking"
	"github.com/yorik8877/meanwhile/internal/tracking/memory"
)

type manualClock struct {
	moment time.Time
}

func (clock *manualClock) Now() time.Time {
	return clock.moment
}

func newTestTracker(t *testing.T) *tracking.Tracker {
	t.Helper()
	return newTestTrackerWithClock(t, tracking.SystemClock{})
}

func newTestTrackerWithClock(t *testing.T, clock tracking.Clock) *tracking.Tracker {
	t.Helper()
	tracker := tracking.NewTracker(memory.NewRepository(), clock, time.UTC)
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

	summaries := tracker.ActiveTaskSummaries()
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

// The list must not be refreshed every second: each refresh builds a new
// template card.
func TestTickUpdatesTimesWithoutRebuildingTheList(t *testing.T) {
	test.NewTempApp(t)
	clock := &manualClock{moment: time.Date(2026, time.October, 19, 10, 0, 0, 0, time.UTC)}
	tracker := newTestTrackerWithClock(t, clock)
	id, err := tracker.CreateTask("Task", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := tracker.Start(id); err != nil {
		t.Fatal(err)
	}
	window := test.NewTempWindow(t, nil)
	tab := newTasksTab(tracker, window)
	window.SetContent(tab.content)
	window.Resize(windowSize)
	createdRows := 0
	createRow := tab.list.CreateItem
	tab.list.CreateItem = func() fyne.CanvasObject {
		createdRows++
		return createRow()
	}

	clock.moment = clock.moment.Add(time.Minute)
	tab.tick()

	if createdRows != 0 {
		t.Errorf("tick created %d list rows, want none", createdRows)
	}
	row := shownRow(t, tab, 0)
	if want := formatDuration(time.Minute); !strings.Contains(row.times.Text, want) {
		t.Errorf("times = %q, want them to contain %s", row.times.Text, want)
	}
}

func TestTickShowsTasksAddedSinceTheLastRefresh(t *testing.T) {
	test.NewTempApp(t)
	tracker := newTestTracker(t)
	window := test.NewTempWindow(t, nil)
	tab := newTasksTab(tracker, window)
	window.SetContent(tab.content)
	window.Resize(windowSize)
	if _, err := tracker.CreateTask("Task", ""); err != nil {
		t.Fatal(err)
	}

	tab.tick()

	if len(tab.summaries) != 1 {
		t.Errorf("tab shows %d tasks after tick, want 1", len(tab.summaries))
	}
}

func shownRow(t *testing.T, tab *tasksTab, index int) *taskRow {
	t.Helper()
	for row, shownIndex := range tab.shownRows {
		if shownIndex == index {
			return row
		}
	}
	t.Fatalf("no row shows task %d", index)
	return nil
}

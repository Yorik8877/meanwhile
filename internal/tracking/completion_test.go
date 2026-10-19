package tracking_test

import (
	"errors"
	"testing"
	"time"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

func taskIDs(summaries []tracking.TaskSummary) []tracking.TaskID {
	ids := make([]tracking.TaskID, len(summaries))
	for index, summary := range summaries {
		ids[index] = summary.Task.ID
	}
	return ids
}

func TestCompletingRunningTaskPausesItAndMovesItToCompleted(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.Start(id))
	fixture.clock.Advance(10 * time.Minute)

	mustSucceed(t, fixture.tracker.Complete(id))

	if fixture.tracker.AnyRunning() {
		t.Error("completed task keeps running")
	}
	if active := fixture.tracker.ActiveTaskSummaries(); len(active) != 0 {
		t.Errorf("active tasks = %v, want none", taskIDs(active))
	}
	completed := fixture.tracker.CompletedTaskSummaries()
	if len(completed) != 1 || completed[0].Total != 10*time.Minute || !completed[0].Task.CompletedAt.Equal(fixture.clock.Now()) {
		t.Errorf("completed = %+v, want the task with 10m completed now", completed)
	}
}

func TestCompletedTaskCannotStart(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.Complete(id))

	for name, start := range map[string]func(tracking.TaskID) error{
		"Start":  fixture.tracker.Start,
		"Toggle": fixture.tracker.Toggle,
	} {
		if err := start(id); !errors.Is(err, tracking.ErrTaskCompleted) {
			t.Errorf("%s: error = %v, want ErrTaskCompleted", name, err)
		}
	}
}

func TestReopenedTaskReturnsToItsPlace(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	oldest := fixture.createTask(t, "Oldest")
	fixture.clock.Advance(time.Second)
	middle := fixture.createTask(t, "Middle")
	fixture.clock.Advance(time.Second)
	newest := fixture.createTask(t, "Newest")
	mustSucceed(t, fixture.tracker.Complete(middle))

	mustSucceed(t, fixture.tracker.Reopen(middle))

	got := taskIDs(fixture.tracker.ActiveTaskSummaries())
	want := []tracking.TaskID{newest, middle, oldest}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("active order = %v, want %v", got, want)
	}
	if completed := fixture.tracker.CompletedTaskSummaries(); len(completed) != 0 {
		t.Errorf("completed = %v, want none", taskIDs(completed))
	}
}

func TestCompletedTasksAreListedRecentlyCompletedFirst(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	first := fixture.createTask(t, "First")
	second := fixture.createTask(t, "Second")
	mustSucceed(t, fixture.tracker.Complete(second))
	fixture.clock.Advance(time.Minute)
	mustSucceed(t, fixture.tracker.Complete(first))

	got := taskIDs(fixture.tracker.CompletedTaskSummaries())

	if len(got) != 2 || got[0] != first || got[1] != second {
		t.Errorf("completed order = %v, want first (completed later) then second", got)
	}
}

func TestCompletedTasksLeaveTrayButStayInHistory(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	morning := time.Date(2026, time.October, 14, 8, 0, 0, 0, time.UTC)
	mustSucceed(t, fixture.tracker.AddInterval(id, morning, morning.Add(time.Hour)))

	mustSucceed(t, fixture.tracker.Complete(id))

	if trayTasks := fixture.tracker.TrayTasks(10); len(trayTasks) != 0 {
		t.Errorf("tray tasks = %+v, want none", trayTasks)
	}
	totals := fixture.tracker.DailyTotals(tracking.DateOf(morning), 1)
	if len(totals) != 1 || totals[0].Total != time.Hour {
		t.Errorf("history = %+v, want 1h on the day", totals)
	}
}

func TestCompletingTwiceKeepsFirstCompletionMoment(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.Complete(id))
	completedAt := fixture.clock.Now()
	fixture.clock.Advance(time.Hour)

	mustSucceed(t, fixture.tracker.Complete(id))

	if got := fixture.tracker.CompletedTaskSummaries()[0].Task.CompletedAt; !got.Equal(completedAt) {
		t.Errorf("completed at = %v, want %v", got, completedAt)
	}
}

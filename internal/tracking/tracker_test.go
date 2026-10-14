package tracking_test

import (
	"errors"
	"testing"
	"time"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

func TestCreateTaskTrimsNameAndRejectsEmptyName(t *testing.T) {
	fixture := newFixture(t, time.UTC)

	id := fixture.createTask(t, "  Review PR  ")
	if name := fixture.summaryOf(t, id).Task.Name; name != "Review PR" {
		t.Errorf("name = %q, want %q", name, "Review PR")
	}

	_, err := fixture.tracker.CreateTask("   ", "note")
	if !errors.Is(err, tracking.ErrEmptyName) {
		t.Errorf("error = %v, want ErrEmptyName", err)
	}
}

func TestStartAndPauseAccumulateTime(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Write report")

	mustSucceed(t, fixture.tracker.Start(id))
	fixture.clock.Advance(30 * time.Minute)
	if summary := fixture.summaryOf(t, id); !summary.Running || summary.Total != 30*time.Minute {
		t.Fatalf("while running: %+v, want running with 30m", summary)
	}

	mustSucceed(t, fixture.tracker.Pause(id))
	fixture.clock.Advance(time.Hour)
	mustSucceed(t, fixture.tracker.Start(id))
	fixture.clock.Advance(15 * time.Minute)
	mustSucceed(t, fixture.tracker.Pause(id))

	summary := fixture.summaryOf(t, id)
	if summary.Running || summary.Total != 45*time.Minute || summary.Today != 45*time.Minute {
		t.Errorf("summary = %+v, want paused with 45m total and today", summary)
	}
}

func TestStartTwiceKeepsSingleInterval(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")

	mustSucceed(t, fixture.tracker.Start(id))
	fixture.clock.Advance(time.Minute)
	mustSucceed(t, fixture.tracker.Start(id))

	if count := len(fixture.tracker.Intervals(id)); count != 1 {
		t.Errorf("interval count = %d, want 1", count)
	}
}

func TestParallelTasksCountIndependently(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	build := fixture.createTask(t, "Build")
	review := fixture.createTask(t, "Review")

	mustSucceed(t, fixture.tracker.Start(build))
	fixture.clock.Advance(10 * time.Minute)
	mustSucceed(t, fixture.tracker.Start(review))
	fixture.clock.Advance(20 * time.Minute)
	mustSucceed(t, fixture.tracker.Pause(build))

	if total := fixture.summaryOf(t, build).Total; total != 30*time.Minute {
		t.Errorf("build total = %v, want 30m", total)
	}
	reviewSummary := fixture.summaryOf(t, review)
	if !reviewSummary.Running || reviewSummary.Total != 20*time.Minute {
		t.Errorf("review = %+v, want running with 20m", reviewSummary)
	}
}

func TestPauseAllStopsEveryRunningTaskAtGivenMoment(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	first := fixture.createTask(t, "First")
	second := fixture.createTask(t, "Second")
	mustSucceed(t, fixture.tracker.Start(first))
	mustSucceed(t, fixture.tracker.Start(second))
	sleepMoment := fixture.clock.Now().Add(5 * time.Minute)
	fixture.clock.Advance(2 * time.Hour)

	mustSucceed(t, fixture.tracker.PauseAll(sleepMoment))

	if fixture.tracker.AnyRunning() {
		t.Fatal("tasks still running after PauseAll")
	}
	for _, id := range []tracking.TaskID{first, second} {
		if total := fixture.summaryOf(t, id).Total; total != 5*time.Minute {
			t.Errorf("task %d total = %v, want 5m", id, total)
		}
	}
}

func TestPauseAfterClockMovedBackGivesZeroLength(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.Start(id))

	fixture.clock.Advance(-10 * time.Minute)
	mustSucceed(t, fixture.tracker.Pause(id))

	if total := fixture.summaryOf(t, id).Total; total != 0 {
		t.Errorf("total = %v, want 0", total)
	}
}

func TestDeleteTaskRemovesItsIntervals(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.Start(id))

	mustSucceed(t, fixture.tracker.DeleteTask(id))

	if len(fixture.tracker.TaskSummaries()) != 0 || fixture.tracker.AnyRunning() {
		t.Error("task or its running interval survived deletion")
	}
	stored, _ := fixture.repository.Intervals()
	if len(stored) != 0 {
		t.Errorf("stored intervals = %d, want 0", len(stored))
	}
}

func TestUpdateTaskChangesNameAndNote(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Old")

	mustSucceed(t, fixture.tracker.UpdateTask(id, " New ", "line one\nline two"))

	task := fixture.summaryOf(t, id).Task
	if task.Name != "New" || task.Note != "line one\nline two" {
		t.Errorf("task = %+v", task)
	}
	if err := fixture.tracker.UpdateTask(id, "", ""); !errors.Is(err, tracking.ErrEmptyName) {
		t.Errorf("error = %v, want ErrEmptyName", err)
	}
}

func TestFailedWriteLeavesStateUnchanged(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	writeError := errors.New("disk full")
	fixture.repository.WriteError = writeError

	err := fixture.tracker.Start(id)

	if !errors.Is(err, writeError) {
		t.Fatalf("error = %v, want wrapped disk full", err)
	}
	if fixture.tracker.AnyRunning() {
		t.Error("task started although the write failed")
	}
}

func TestListenersAreNotifiedAfterSuccessfulChangesOnly(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	notifications := 0
	fixture.tracker.Subscribe(func() {
		notifications++
		fixture.tracker.TaskSummaries()
	})

	fixture.createTask(t, "Task")
	_, _ = fixture.tracker.CreateTask("", "")

	if notifications != 1 {
		t.Errorf("notifications = %d, want 1", notifications)
	}
}

func TestTaskSummariesListNewestFirst(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	older := fixture.createTask(t, "Older")
	fixture.clock.Advance(time.Second)
	newer := fixture.createTask(t, "Newer")

	summaries := fixture.tracker.TaskSummaries()

	if len(summaries) != 2 || summaries[0].Task.ID != newer || summaries[1].Task.ID != older {
		t.Errorf("order = %+v, want newer then older", summaries)
	}
}

func TestTodayCountsOnlyThePartAfterMidnight(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Night shift")
	fixture.clock.moment = time.Date(2026, time.October, 13, 23, 0, 0, 0, time.UTC)
	mustSucceed(t, fixture.tracker.Start(id))

	fixture.clock.moment = time.Date(2026, time.October, 14, 1, 30, 0, 0, time.UTC)

	summary := fixture.summaryOf(t, id)
	if summary.Today != 90*time.Minute || summary.Total != 150*time.Minute {
		t.Errorf("today = %v, total = %v, want 1h30m and 2h30m", summary.Today, summary.Total)
	}
}

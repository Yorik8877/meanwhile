package tracking_test

import (
	"testing"
	"time"

	"stopwatch/internal/tracking"
	"stopwatch/internal/tracking/memory"
)

type manualClock struct {
	moment time.Time
}

func (clock *manualClock) Now() time.Time {
	return clock.moment
}

func (clock *manualClock) Advance(duration time.Duration) {
	clock.moment = clock.moment.Add(duration)
}

type fixture struct {
	tracker    *tracking.Tracker
	clock      *manualClock
	repository *memory.Repository
}

// newFixture starts the clock at 2026-10-14 10:00 in the given location.
func newFixture(t *testing.T, location *time.Location) fixture {
	t.Helper()
	repository := memory.NewRepository()
	clock := &manualClock{moment: time.Date(2026, time.October, 14, 10, 0, 0, 0, location)}
	tracker := tracking.NewTracker(repository, clock, location)
	if err := tracker.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	return fixture{tracker: tracker, clock: clock, repository: repository}
}

func (fixture fixture) createTask(t *testing.T, name string) tracking.TaskID {
	t.Helper()
	id, err := fixture.tracker.CreateTask(name, "")
	if err != nil {
		t.Fatalf("create task %q: %v", name, err)
	}
	return id
}

func (fixture fixture) summaryOf(t *testing.T, id tracking.TaskID) tracking.TaskSummary {
	t.Helper()
	for _, summary := range fixture.tracker.TaskSummaries() {
		if summary.Task.ID == id {
			return summary
		}
	}
	t.Fatalf("task %d not found in summaries", id)
	return tracking.TaskSummary{}
}

func mustSucceed(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

package tracking_test

import (
	"errors"
	"testing"
	"time"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

// The fixture clock is 2026-10-14 10:00 UTC.
func at(hour, minute int) time.Time {
	return time.Date(2026, time.October, 14, hour, minute, 0, 0, time.UTC)
}

func TestAddIntervalValidatesBounds(t *testing.T) {
	cases := []struct {
		name       string
		start, end time.Time
		want       error
	}{
		{"valid", at(8, 0), at(9, 0), nil},
		{"end equals start", at(8, 0), at(8, 0), tracking.ErrEndBeforeStart},
		{"end before start", at(9, 0), at(8, 0), tracking.ErrEndBeforeStart},
		{"end in the future", at(9, 0), at(11, 0), tracking.ErrIntervalInFuture},
		{"missing end", at(9, 0), time.Time{}, tracking.ErrMissingEnd},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newFixture(t, time.UTC)
			id := fixture.createTask(t, "Task")

			err := fixture.tracker.AddInterval(id, testCase.start, testCase.end)

			if !errors.Is(err, testCase.want) {
				t.Errorf("error = %v, want %v", err, testCase.want)
			}
		})
	}
}

func TestAddIntervalRejectsOverlapWithinTaskOnly(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	first := fixture.createTask(t, "First")
	second := fixture.createTask(t, "Second")
	mustSucceed(t, fixture.tracker.AddInterval(first, at(8, 0), at(9, 0)))

	if err := fixture.tracker.AddInterval(first, at(8, 30), at(9, 30)); !errors.Is(err, tracking.ErrIntervalsOverlap) {
		t.Errorf("overlap in same task: error = %v, want ErrIntervalsOverlap", err)
	}
	if err := fixture.tracker.AddInterval(first, at(9, 0), at(9, 30)); err != nil {
		t.Errorf("adjacent interval: error = %v, want nil", err)
	}
	if err := fixture.tracker.AddInterval(second, at(8, 30), at(9, 30)); err != nil {
		t.Errorf("overlap with another task: error = %v, want nil", err)
	}
}

func TestAddIntervalRejectsOverlapWithRunningInterval(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	fixture.clock.moment = at(9, 0)
	mustSucceed(t, fixture.tracker.Start(id))
	fixture.clock.moment = at(10, 0)

	err := fixture.tracker.AddInterval(id, at(9, 30), at(9, 45))

	if !errors.Is(err, tracking.ErrIntervalsOverlap) {
		t.Errorf("error = %v, want ErrIntervalsOverlap", err)
	}
}

func TestUpdateRunningIntervalChangesOnlyStart(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.Start(id))
	running := fixture.tracker.Intervals(id)[0]

	if err := fixture.tracker.UpdateInterval(running.ID, at(9, 0), at(9, 30)); !errors.Is(err, tracking.ErrRunningIntervalEnd) {
		t.Errorf("setting end: error = %v, want ErrRunningIntervalEnd", err)
	}
	if err := fixture.tracker.UpdateInterval(running.ID, at(10, 5), time.Time{}); !errors.Is(err, tracking.ErrIntervalInFuture) {
		t.Errorf("future start: error = %v, want ErrIntervalInFuture", err)
	}
	mustSucceed(t, fixture.tracker.UpdateInterval(running.ID, at(9, 0), time.Time{}))

	if summary := fixture.summaryOf(t, id); !summary.Running || summary.Total != time.Hour {
		t.Errorf("summary = %+v, want running with 1h", summary)
	}
}

func TestUpdateClosedIntervalRequiresEndAndValidates(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.AddInterval(id, at(8, 0), at(9, 0)))
	mustSucceed(t, fixture.tracker.AddInterval(id, at(9, 0), at(9, 30)))
	earlier := fixture.tracker.Intervals(id)[1]

	if err := fixture.tracker.UpdateInterval(earlier.ID, at(8, 0), time.Time{}); !errors.Is(err, tracking.ErrMissingEnd) {
		t.Errorf("missing end: error = %v, want ErrMissingEnd", err)
	}
	if err := fixture.tracker.UpdateInterval(earlier.ID, at(8, 0), at(9, 15)); !errors.Is(err, tracking.ErrIntervalsOverlap) {
		t.Errorf("overlap: error = %v, want ErrIntervalsOverlap", err)
	}
	mustSucceed(t, fixture.tracker.UpdateInterval(earlier.ID, at(7, 0), at(9, 0)))

	if total := fixture.summaryOf(t, id).Total; total != 150*time.Minute {
		t.Errorf("total = %v, want 2h30m", total)
	}
}

func TestDeleteInterval(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.AddInterval(id, at(8, 0), at(9, 0)))
	interval := fixture.tracker.Intervals(id)[0]

	mustSucceed(t, fixture.tracker.DeleteInterval(interval.ID))

	if total := fixture.summaryOf(t, id).Total; total != 0 {
		t.Errorf("total = %v, want 0", total)
	}
	if err := fixture.tracker.DeleteInterval(interval.ID); !errors.Is(err, tracking.ErrIntervalNotFound) {
		t.Errorf("second delete: error = %v, want ErrIntervalNotFound", err)
	}
}

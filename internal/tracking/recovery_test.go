package tracking_test

import (
	"testing"
	"time"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

func TestRecoverAfterCrashClosesIntervalsAtLastHeartbeat(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.Start(id))
	fixture.clock.Advance(20 * time.Minute)
	mustSucceed(t, fixture.tracker.RecordHeartbeat())

	restarted := tracking.NewTracker(fixture.repository, fixture.clock, time.UTC)
	mustSucceed(t, restarted.Load())
	fixture.clock.Advance(5 * time.Hour)
	mustSucceed(t, restarted.RecoverAfterCrash())

	if restarted.AnyRunning() {
		t.Fatal("interval still running after recovery")
	}
	if total := restarted.ActiveTaskSummaries()[0].Total; total != 20*time.Minute {
		t.Errorf("total = %v, want 20m", total)
	}
}

func TestRecoverAfterCrashWithoutUsefulHeartbeatGivesZeroLength(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	mustSucceed(t, fixture.tracker.RecordHeartbeat())
	fixture.clock.Advance(time.Minute)
	id := fixture.createTask(t, "Task")
	mustSucceed(t, fixture.tracker.Start(id))
	fixture.clock.Advance(time.Hour)

	mustSucceed(t, fixture.tracker.RecoverAfterCrash())

	if total := fixture.summaryOf(t, id).Total; total != 0 {
		t.Errorf("total = %v, want 0", total)
	}
}

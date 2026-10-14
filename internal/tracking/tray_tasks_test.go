package tracking_test

import (
	"testing"
	"time"

	"stopwatch/internal/tracking"
)

func TestTrayTasksPutRunningFirstThenRecentlyActive(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	stale := fixture.createTask(t, "Stale")
	fixture.clock.Advance(time.Minute)
	fixture.createTask(t, "Recent")
	fixture.clock.Advance(time.Minute)
	running := fixture.createTask(t, "Running")
	fixture.clock.Advance(time.Minute)
	mustSucceed(t, fixture.tracker.Start(stale))
	mustSucceed(t, fixture.tracker.Pause(stale))
	fixture.clock.Advance(time.Minute)
	mustSucceed(t, fixture.tracker.Start(running))

	trayTasks := fixture.tracker.TrayTasks(2)

	if len(trayTasks) != 2 || trayTasks[0].Task.ID != running || !trayTasks[0].Running || trayTasks[1].Task.ID != stale {
		t.Errorf("tray tasks = %+v, want running, then stale because it was started after Recent was created", trayTasks)
	}
}

func TestTrayTasksNeverHideRunningTasks(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	var ids []tracking.TaskID
	for range 3 {
		id := fixture.createTask(t, "Task")
		mustSucceed(t, fixture.tracker.Start(id))
		ids = append(ids, id)
	}

	if count := len(fixture.tracker.TrayTasks(2)); count != len(ids) {
		t.Errorf("tray task count = %d, want %d", count, len(ids))
	}
}

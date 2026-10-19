package storage_test

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/yorik8877/meanwhile/internal/storage"
	"github.com/yorik8877/meanwhile/internal/tracking"
)

func openStore(t *testing.T, path string) *storage.Store {
	t.Helper()
	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func mustSucceed(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

var morning = time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)

func TestDataSurvivesReopening(t *testing.T) {
	path := filepath.Join(t.TempDir(), "my meanwhile ä.db")
	store := openStore(t, path)
	taskID, err := store.InsertTask(tracking.Task{Name: "Task", Note: "note", CreatedAt: morning})
	mustSucceed(t, err)
	_, err = store.InsertInterval(tracking.Interval{TaskID: taskID, StartedAt: morning, StoppedAt: morning.Add(time.Hour)})
	mustSucceed(t, err)
	_, err = store.InsertInterval(tracking.Interval{TaskID: taskID, StartedAt: morning.Add(2 * time.Hour)})
	mustSucceed(t, err)
	mustSucceed(t, store.SaveHeartbeat(morning.Add(3*time.Hour)))
	mustSucceed(t, store.Close())

	reopened := openStore(t, path)
	tasks, err := reopened.Tasks()
	mustSucceed(t, err)
	intervals, err := reopened.Intervals()
	mustSucceed(t, err)
	heartbeat, err := reopened.Heartbeat()
	mustSucceed(t, err)

	if len(tasks) != 1 || tasks[0].Name != "Task" || tasks[0].Note != "note" || !tasks[0].CreatedAt.Equal(morning) {
		t.Errorf("tasks = %+v", tasks)
	}
	if len(intervals) != 2 || !intervals[0].StoppedAt.Equal(morning.Add(time.Hour)) || !intervals[1].IsRunning() {
		t.Errorf("intervals = %+v", intervals)
	}
	if !heartbeat.Equal(morning.Add(3 * time.Hour)) {
		t.Errorf("heartbeat = %v", heartbeat)
	}
}

func TestHeartbeatIsZeroWhenNeverSaved(t *testing.T) {
	store := openStore(t, filepath.Join(t.TempDir(), "meanwhile.db"))

	heartbeat, err := store.Heartbeat()

	mustSucceed(t, err)
	if !heartbeat.IsZero() {
		t.Errorf("heartbeat = %v, want zero", heartbeat)
	}
}

func TestSecondRunningIntervalOfTaskIsRejected(t *testing.T) {
	store := openStore(t, filepath.Join(t.TempDir(), "meanwhile.db"))
	taskID, err := store.InsertTask(tracking.Task{Name: "Task", CreatedAt: morning})
	mustSucceed(t, err)
	_, err = store.InsertInterval(tracking.Interval{TaskID: taskID, StartedAt: morning})
	mustSucceed(t, err)

	_, err = store.InsertInterval(tracking.Interval{TaskID: taskID, StartedAt: morning.Add(time.Minute)})

	if err == nil {
		t.Error("second running interval was accepted")
	}
}

func TestDeletingTaskDeletesItsIntervals(t *testing.T) {
	store := openStore(t, filepath.Join(t.TempDir(), "meanwhile.db"))
	taskID, err := store.InsertTask(tracking.Task{Name: "Task", CreatedAt: morning})
	mustSucceed(t, err)
	_, err = store.InsertInterval(tracking.Interval{TaskID: taskID, StartedAt: morning, StoppedAt: morning.Add(time.Hour)})
	mustSucceed(t, err)

	mustSucceed(t, store.DeleteTask(taskID))

	intervals, err := store.Intervals()
	mustSucceed(t, err)
	if len(intervals) != 0 {
		t.Errorf("intervals = %+v, want none", intervals)
	}
}

func TestUpdateIntervalsIsAllOrNothing(t *testing.T) {
	store := openStore(t, filepath.Join(t.TempDir(), "meanwhile.db"))
	taskID, err := store.InsertTask(tracking.Task{Name: "Task", CreatedAt: morning})
	mustSucceed(t, err)
	intervalID, err := store.InsertInterval(tracking.Interval{TaskID: taskID, StartedAt: morning})
	mustSucceed(t, err)

	err = store.UpdateIntervals([]tracking.Interval{
		{ID: intervalID, TaskID: taskID, StartedAt: morning, StoppedAt: morning.Add(time.Hour)},
		{ID: intervalID + 100, TaskID: taskID, StartedAt: morning, StoppedAt: morning.Add(time.Hour)},
	})

	if err == nil {
		t.Fatal("update of a missing interval succeeded")
	}
	intervals, err := store.Intervals()
	mustSucceed(t, err)
	if !intervals[0].IsRunning() {
		t.Errorf("first interval was closed although the transaction failed")
	}
}

func TestEndBeforeStartIsRejectedByDatabase(t *testing.T) {
	store := openStore(t, filepath.Join(t.TempDir(), "meanwhile.db"))
	taskID, err := store.InsertTask(tracking.Task{Name: "Task", CreatedAt: morning})
	mustSucceed(t, err)

	_, err = store.InsertInterval(tracking.Interval{TaskID: taskID, StartedAt: morning, StoppedAt: morning.Add(-time.Minute)})

	if err == nil {
		t.Error("interval ending before its start was accepted")
	}
}

func TestLanguageIsEmptyUntilSavedAndSurvivesReopening(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meanwhile.db")
	store := openStore(t, path)

	language, err := store.Language()
	mustSucceed(t, err)
	if language != "" {
		t.Errorf("language = %q, want empty", language)
	}

	mustSucceed(t, store.SaveLanguage("ru"))
	mustSucceed(t, store.Close())

	language, err = openStore(t, path).Language()
	mustSucceed(t, err)
	if language != "ru" {
		t.Errorf("language after reopening = %q, want ru", language)
	}
}

func TestCompletionSurvivesReopening(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meanwhile.db")
	store := openStore(t, path)
	taskID, err := store.InsertTask(tracking.Task{Name: "Task", CreatedAt: morning})
	mustSucceed(t, err)
	mustSucceed(t, store.UpdateTask(tracking.Task{ID: taskID, Name: "Task", CompletedAt: morning.Add(time.Hour)}))
	mustSucceed(t, store.Close())

	tasks, err := openStore(t, path).Tasks()

	mustSucceed(t, err)
	if !tasks[0].CompletedAt.Equal(morning.Add(time.Hour)) {
		t.Errorf("completed at = %v, want %v", tasks[0].CompletedAt, morning.Add(time.Hour))
	}
}

// A database created by the first release must open with its tasks intact
// and with every task still in progress.
func TestFirstSchemaVersionIsUpgraded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meanwhile.db")
	database, err := sql.Open("sqlite", path)
	mustSucceed(t, err)
	_, err = database.Exec(firstSchemaVersion + `
		INSERT INTO tasks (id, name, note, created_at) VALUES (1, 'Old task', 'note', 1000);
		PRAGMA user_version = 1;`)
	mustSucceed(t, err)
	mustSucceed(t, database.Close())

	tasks, err := openStore(t, path).Tasks()

	mustSucceed(t, err)
	if len(tasks) != 1 || tasks[0].Name != "Old task" || !tasks[0].CompletedAt.IsZero() {
		t.Errorf("tasks = %+v, want the old task in progress", tasks)
	}
}

const firstSchemaVersion = `
	CREATE TABLE tasks (
		id         INTEGER PRIMARY KEY,
		name       TEXT    NOT NULL,
		note       TEXT    NOT NULL DEFAULT '',
		created_at INTEGER NOT NULL
	);
	CREATE TABLE intervals (
		id         INTEGER PRIMARY KEY,
		task_id    INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		started_at INTEGER NOT NULL,
		stopped_at INTEGER CHECK (stopped_at IS NULL OR stopped_at >= started_at)
	);
	CREATE TABLE app_state (
		key   TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);`

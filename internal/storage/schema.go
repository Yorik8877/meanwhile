package storage

// migrations[i] upgrades the schema from version i to version i+1. The
// current version is kept in PRAGMA user_version.
var migrations = []string{
	`CREATE TABLE tasks (
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
	CREATE INDEX intervals_by_task ON intervals(task_id);
	CREATE UNIQUE INDEX one_running_interval_per_task ON intervals(task_id) WHERE stopped_at IS NULL;
	CREATE TABLE app_state (
		key   TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);`,
}

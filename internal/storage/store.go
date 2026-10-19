// Package storage keeps tracking data in an SQLite database.
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "modernc.org/sqlite"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

const (
	heartbeatKey = "last_heartbeat_at"
	languageKey  = "language"
)

// Store implements tracking.Repository. Moments are stored as Unix
// milliseconds.
type Store struct {
	database *sql.DB
}

var _ tracking.Repository = (*Store)(nil)

// Open opens or creates the database file and brings its schema up to date.
func Open(path string) (*Store, error) {
	source := path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	database, err := sql.Open("sqlite", source)
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", path, err)
	}
	// SQLite allows one writer; a single connection also keeps the
	// per-connection pragmas in effect for every statement.
	database.SetMaxOpenConns(1)

	store := &Store{database: database}
	if err := store.migrate(); err != nil {
		database.Close()
		return nil, fmt.Errorf("migrate database %s: %w", path, err)
	}
	return store, nil
}

func (store *Store) Close() error {
	return store.database.Close()
}

func (store *Store) migrate() error {
	var version int
	if err := store.database.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	for nextVersion := version; nextVersion < len(migrations); nextVersion++ {
		if err := store.applyMigration(nextVersion); err != nil {
			return err
		}
	}
	return nil
}

func (store *Store) applyMigration(fromVersion int) error {
	transaction, err := store.database.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	if _, err := transaction.Exec(migrations[fromVersion]); err != nil {
		return fmt.Errorf("apply migration %d: %w", fromVersion+1, err)
	}
	if _, err := transaction.Exec(fmt.Sprintf("PRAGMA user_version = %d", fromVersion+1)); err != nil {
		return fmt.Errorf("set schema version %d: %w", fromVersion+1, err)
	}
	return transaction.Commit()
}

func (store *Store) Tasks() ([]tracking.Task, error) {
	rows, err := store.database.Query("SELECT id, name, note, created_at, completed_at FROM tasks ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []tracking.Task
	for rows.Next() {
		var task tracking.Task
		var createdAt int64
		var completedAt sql.NullInt64
		if err := rows.Scan(&task.ID, &task.Name, &task.Note, &createdAt, &completedAt); err != nil {
			return nil, err
		}
		task.CreatedAt = fromMilliseconds(createdAt)
		if completedAt.Valid {
			task.CompletedAt = fromMilliseconds(completedAt.Int64)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (store *Store) Intervals() ([]tracking.Interval, error) {
	rows, err := store.database.Query("SELECT id, task_id, started_at, stopped_at FROM intervals ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var intervals []tracking.Interval
	for rows.Next() {
		var interval tracking.Interval
		var startedAt int64
		var stoppedAt sql.NullInt64
		if err := rows.Scan(&interval.ID, &interval.TaskID, &startedAt, &stoppedAt); err != nil {
			return nil, err
		}
		interval.StartedAt = fromMilliseconds(startedAt)
		if stoppedAt.Valid {
			interval.StoppedAt = fromMilliseconds(stoppedAt.Int64)
		}
		intervals = append(intervals, interval)
	}
	return intervals, rows.Err()
}

func (store *Store) InsertTask(task tracking.Task) (tracking.TaskID, error) {
	result, err := store.database.Exec(
		"INSERT INTO tasks (name, note, created_at, completed_at) VALUES (?, ?, ?, ?)",
		task.Name, task.Note, task.CreatedAt.UnixMilli(), nullableMilliseconds(task.CompletedAt),
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return tracking.TaskID(id), err
}

func (store *Store) UpdateTask(task tracking.Task) error {
	result, err := store.database.Exec(
		"UPDATE tasks SET name = ?, note = ?, completed_at = ? WHERE id = ?",
		task.Name, task.Note, nullableMilliseconds(task.CompletedAt), task.ID,
	)
	if err != nil {
		return err
	}
	return requireOneRow(result)
}

func (store *Store) DeleteTask(id tracking.TaskID) error {
	_, err := store.database.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func (store *Store) InsertInterval(interval tracking.Interval) (tracking.IntervalID, error) {
	result, err := store.database.Exec(
		"INSERT INTO intervals (task_id, started_at, stopped_at) VALUES (?, ?, ?)",
		interval.TaskID, interval.StartedAt.UnixMilli(), nullableMilliseconds(interval.StoppedAt),
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return tracking.IntervalID(id), err
}

// UpdateIntervals writes all intervals in one transaction, so that pausing
// several tasks at once either fully succeeds or changes nothing.
func (store *Store) UpdateIntervals(intervals []tracking.Interval) error {
	transaction, err := store.database.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	for _, interval := range intervals {
		result, err := transaction.Exec(
			"UPDATE intervals SET started_at = ?, stopped_at = ? WHERE id = ?",
			interval.StartedAt.UnixMilli(), nullableMilliseconds(interval.StoppedAt), interval.ID,
		)
		if err != nil {
			return err
		}
		if err := requireOneRow(result); err != nil {
			return err
		}
	}
	return transaction.Commit()
}

func (store *Store) DeleteInterval(id tracking.IntervalID) error {
	_, err := store.database.Exec("DELETE FROM intervals WHERE id = ?", id)
	return err
}

func (store *Store) Heartbeat() (time.Time, error) {
	value, err := store.stateValue(heartbeatKey)
	if err != nil || value == "" {
		return time.Time{}, err
	}
	milliseconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse heartbeat %q: %w", value, err)
	}
	return fromMilliseconds(milliseconds), nil
}

func (store *Store) SaveHeartbeat(moment time.Time) error {
	return store.saveStateValue(heartbeatKey, strconv.FormatInt(moment.UnixMilli(), 10))
}

// Language returns the chosen interface language code, or an empty string
// when the system language is used.
func (store *Store) Language() (string, error) {
	return store.stateValue(languageKey)
}

func (store *Store) SaveLanguage(code string) error {
	return store.saveStateValue(languageKey, code)
}

// stateValue returns an empty string for a key that was never saved.
func (store *Store) stateValue(key string) (string, error) {
	var value string
	err := store.database.QueryRow("SELECT value FROM app_state WHERE key = ?", key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

func (store *Store) saveStateValue(key, value string) error {
	_, err := store.database.Exec(
		"INSERT INTO app_state (key, value) VALUES (?, ?) ON CONFLICT (key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	return err
}

func requireOneRow(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("expected to change 1 row, changed %d", affected)
	}
	return nil
}

func fromMilliseconds(milliseconds int64) time.Time {
	return time.UnixMilli(milliseconds)
}

func nullableMilliseconds(moment time.Time) sql.NullInt64 {
	if moment.IsZero() {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: moment.UnixMilli(), Valid: true}
}

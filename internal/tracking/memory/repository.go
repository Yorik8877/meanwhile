// Package memory provides an in-memory tracking.Repository for tests.
package memory

import (
	"cmp"
	"errors"
	"slices"
	"sync"
	"time"

	"stopwatch/internal/tracking"
)

var errNotFound = errors.New("not found")

type Repository struct {
	// WriteError, when set, is returned by every write instead of applying it.
	WriteError error

	mutex          sync.Mutex
	nextTaskID     tracking.TaskID
	nextIntervalID tracking.IntervalID
	tasks          map[tracking.TaskID]tracking.Task
	intervals      map[tracking.IntervalID]tracking.Interval
	heartbeat      time.Time
}

func NewRepository() *Repository {
	return &Repository{
		tasks:     map[tracking.TaskID]tracking.Task{},
		intervals: map[tracking.IntervalID]tracking.Interval{},
	}
}

func (repository *Repository) Tasks() ([]tracking.Task, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	tasks := make([]tracking.Task, 0, len(repository.tasks))
	for _, task := range repository.tasks {
		tasks = append(tasks, task)
	}
	slices.SortFunc(tasks, func(first, second tracking.Task) int { return cmp.Compare(first.ID, second.ID) })
	return tasks, nil
}

func (repository *Repository) Intervals() ([]tracking.Interval, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	intervals := make([]tracking.Interval, 0, len(repository.intervals))
	for _, interval := range repository.intervals {
		intervals = append(intervals, interval)
	}
	slices.SortFunc(intervals, func(first, second tracking.Interval) int { return cmp.Compare(first.ID, second.ID) })
	return intervals, nil
}

func (repository *Repository) InsertTask(task tracking.Task) (tracking.TaskID, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.WriteError != nil {
		return 0, repository.WriteError
	}
	repository.nextTaskID++
	task.ID = repository.nextTaskID
	repository.tasks[task.ID] = task
	return task.ID, nil
}

func (repository *Repository) UpdateTask(task tracking.Task) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.WriteError != nil {
		return repository.WriteError
	}
	if _, found := repository.tasks[task.ID]; !found {
		return errNotFound
	}
	repository.tasks[task.ID] = task
	return nil
}

func (repository *Repository) DeleteTask(id tracking.TaskID) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.WriteError != nil {
		return repository.WriteError
	}
	delete(repository.tasks, id)
	for intervalID, interval := range repository.intervals {
		if interval.TaskID == id {
			delete(repository.intervals, intervalID)
		}
	}
	return nil
}

func (repository *Repository) InsertInterval(interval tracking.Interval) (tracking.IntervalID, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.WriteError != nil {
		return 0, repository.WriteError
	}
	repository.nextIntervalID++
	interval.ID = repository.nextIntervalID
	repository.intervals[interval.ID] = interval
	return interval.ID, nil
}

func (repository *Repository) UpdateIntervals(intervals []tracking.Interval) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.WriteError != nil {
		return repository.WriteError
	}
	for _, interval := range intervals {
		if _, found := repository.intervals[interval.ID]; !found {
			return errNotFound
		}
	}
	for _, interval := range intervals {
		repository.intervals[interval.ID] = interval
	}
	return nil
}

func (repository *Repository) DeleteInterval(id tracking.IntervalID) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.WriteError != nil {
		return repository.WriteError
	}
	delete(repository.intervals, id)
	return nil
}

func (repository *Repository) Heartbeat() (time.Time, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	return repository.heartbeat, nil
}

func (repository *Repository) SaveHeartbeat(moment time.Time) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.WriteError != nil {
		return repository.WriteError
	}
	repository.heartbeat = moment
	return nil
}

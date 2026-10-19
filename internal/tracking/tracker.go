package tracking

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

// storedPrecision matches the millisecond precision of the database, so that
// values in memory and in storage are always equal.
const storedPrecision = time.Millisecond

// Tracker is the single entry point for reading and changing tasks and
// intervals. Every change is written to the repository first and applied to
// memory only after the write succeeds. It is safe for concurrent use.
type Tracker struct {
	mutex      sync.Mutex
	repository Repository
	clock      Clock
	location   *time.Location
	tasks      map[TaskID]Task
	intervals  map[IntervalID]Interval
	listeners  []func()
}

func NewTracker(repository Repository, clock Clock, location *time.Location) *Tracker {
	return &Tracker{
		repository: repository,
		clock:      clock,
		location:   location,
		tasks:      map[TaskID]Task{},
		intervals:  map[IntervalID]Interval{},
	}
}

func (tracker *Tracker) Load() error {
	tasks, err := tracker.repository.Tasks()
	if err != nil {
		return fmt.Errorf("load tasks: %w", err)
	}
	intervals, err := tracker.repository.Intervals()
	if err != nil {
		return fmt.Errorf("load intervals: %w", err)
	}

	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	for _, task := range tasks {
		tracker.tasks[task.ID] = task
	}
	for _, interval := range intervals {
		tracker.intervals[interval.ID] = interval
	}
	return nil
}

// Subscribe registers a listener that is called after every successful
// change, on the goroutine that made the change.
func (tracker *Tracker) Subscribe(listener func()) {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	tracker.listeners = append(tracker.listeners, listener)
}

// ValidateTaskName returns the name without surrounding spaces, or
// ErrEmptyName when nothing is left.
func ValidateTaskName(name string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", ErrEmptyName
	}
	return trimmedName, nil
}

func (tracker *Tracker) CreateTask(name, note string) (TaskID, error) {
	trimmedName, err := ValidateTaskName(name)
	if err != nil {
		return 0, err
	}

	var createdID TaskID
	err = tracker.change(func() error {
		task := Task{Name: trimmedName, Note: note, CreatedAt: tracker.now()}
		id, err := tracker.repository.InsertTask(task)
		if err != nil {
			return fmt.Errorf("insert task: %w", err)
		}
		task.ID = id
		tracker.tasks[id] = task
		createdID = id
		return nil
	})
	return createdID, err
}

func (tracker *Tracker) UpdateTask(id TaskID, name, note string) error {
	trimmedName, err := ValidateTaskName(name)
	if err != nil {
		return err
	}

	return tracker.change(func() error {
		task, found := tracker.tasks[id]
		if !found {
			return ErrTaskNotFound
		}
		task.Name = trimmedName
		task.Note = note
		return tracker.storeTask(task)
	})
}

func (tracker *Tracker) DeleteTask(id TaskID) error {
	return tracker.change(func() error {
		if _, found := tracker.tasks[id]; !found {
			return ErrTaskNotFound
		}
		if err := tracker.repository.DeleteTask(id); err != nil {
			return fmt.Errorf("delete task: %w", err)
		}
		delete(tracker.tasks, id)
		for intervalID, interval := range tracker.intervals {
			if interval.TaskID == id {
				delete(tracker.intervals, intervalID)
			}
		}
		return nil
	})
}

// Start does nothing when the task is already running. A completed task
// has to be reopened first.
func (tracker *Tracker) Start(id TaskID) error {
	return tracker.change(func() error {
		task, found := tracker.tasks[id]
		if !found {
			return ErrTaskNotFound
		}
		if task.IsCompleted() {
			return ErrTaskCompleted
		}
		if _, running := tracker.runningInterval(id); running {
			return nil
		}
		interval := Interval{TaskID: id, StartedAt: tracker.now()}
		intervalID, err := tracker.repository.InsertInterval(interval)
		if err != nil {
			return fmt.Errorf("insert interval: %w", err)
		}
		interval.ID = intervalID
		tracker.intervals[intervalID] = interval
		return nil
	})
}

// Pause does nothing when the task is not running.
func (tracker *Tracker) Pause(id TaskID) error {
	return tracker.change(func() error {
		if _, found := tracker.tasks[id]; !found {
			return ErrTaskNotFound
		}
		interval, running := tracker.runningInterval(id)
		if !running {
			return nil
		}
		return tracker.storeIntervals([]Interval{interval.closedAt(tracker.now())})
	})
}

func (tracker *Tracker) Toggle(id TaskID) error {
	tracker.mutex.Lock()
	_, running := tracker.runningInterval(id)
	tracker.mutex.Unlock()

	if running {
		return tracker.Pause(id)
	}
	return tracker.Start(id)
}

// PauseAll closes every running interval at the given moment, for example
// the moment the computer started going to sleep.
func (tracker *Tracker) PauseAll(moment time.Time) error {
	return tracker.change(func() error {
		var closed []Interval
		for _, interval := range tracker.intervals {
			if interval.IsRunning() {
				closed = append(closed, interval.closedAt(moment.Truncate(storedPrecision)))
			}
		}
		if len(closed) == 0 {
			return nil
		}
		return tracker.storeIntervals(closed)
	})
}

// change runs the mutation under the lock and notifies listeners outside of
// it, so that listeners may read from the tracker.
func (tracker *Tracker) change(mutation func() error) error {
	tracker.mutex.Lock()
	err := mutation()
	listeners := slices.Clone(tracker.listeners)
	tracker.mutex.Unlock()

	if err != nil {
		return err
	}
	for _, listener := range listeners {
		listener()
	}
	return nil
}

func (tracker *Tracker) storeTask(task Task) error {
	if err := tracker.repository.UpdateTask(task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	tracker.tasks[task.ID] = task
	return nil
}

func (tracker *Tracker) storeIntervals(intervals []Interval) error {
	if err := tracker.repository.UpdateIntervals(intervals); err != nil {
		return fmt.Errorf("update intervals: %w", err)
	}
	for _, interval := range intervals {
		tracker.intervals[interval.ID] = interval
	}
	return nil
}

func (tracker *Tracker) runningInterval(taskID TaskID) (Interval, bool) {
	for _, interval := range tracker.intervals {
		if interval.TaskID == taskID && interval.IsRunning() {
			return interval, true
		}
	}
	return Interval{}, false
}

// now truncation also strips the monotonic clock reading, so that moments
// compare the same way before and after a round trip through storage.
func (tracker *Tracker) now() time.Time {
	return tracker.clock.Now().Truncate(storedPrecision)
}

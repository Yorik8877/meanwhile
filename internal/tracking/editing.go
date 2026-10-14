package tracking

import (
	"fmt"
	"time"
)

// AddInterval records work that was not timed, for example when the timer
// was not started in time.
func (tracker *Tracker) AddInterval(taskID TaskID, start, end time.Time) error {
	if end.IsZero() {
		return ErrMissingEnd
	}

	return tracker.change(func() error {
		if _, found := tracker.tasks[taskID]; !found {
			return ErrTaskNotFound
		}
		candidate := Interval{
			TaskID:    taskID,
			StartedAt: start.Truncate(storedPrecision),
			StoppedAt: end.Truncate(storedPrecision),
		}
		if err := tracker.validate(candidate); err != nil {
			return err
		}
		id, err := tracker.repository.InsertInterval(candidate)
		if err != nil {
			return fmt.Errorf("insert interval: %w", err)
		}
		candidate.ID = id
		tracker.intervals[id] = candidate
		return nil
	})
}

// UpdateInterval changes the bounds of an interval. A running interval keeps
// running, so its end must be zero.
func (tracker *Tracker) UpdateInterval(id IntervalID, start, end time.Time) error {
	return tracker.change(func() error {
		existing, found := tracker.intervals[id]
		if !found {
			return ErrIntervalNotFound
		}
		if existing.IsRunning() && !end.IsZero() {
			return ErrRunningIntervalEnd
		}
		if !existing.IsRunning() && end.IsZero() {
			return ErrMissingEnd
		}
		candidate := existing
		candidate.StartedAt = start.Truncate(storedPrecision)
		candidate.StoppedAt = end.Truncate(storedPrecision)
		if err := tracker.validate(candidate); err != nil {
			return err
		}
		return tracker.storeIntervals([]Interval{candidate})
	})
}

func (tracker *Tracker) DeleteInterval(id IntervalID) error {
	return tracker.change(func() error {
		if _, found := tracker.intervals[id]; !found {
			return ErrIntervalNotFound
		}
		if err := tracker.repository.DeleteInterval(id); err != nil {
			return fmt.Errorf("delete interval: %w", err)
		}
		delete(tracker.intervals, id)
		return nil
	})
}

func (tracker *Tracker) validate(candidate Interval) error {
	now := tracker.now()
	if err := validateBounds(candidate, now); err != nil {
		return err
	}
	for _, other := range tracker.intervals {
		if other.ID == candidate.ID || other.TaskID != candidate.TaskID {
			continue
		}
		if candidate.overlaps(other, now) {
			return ErrIntervalsOverlap
		}
	}
	return nil
}

func validateBounds(candidate Interval, now time.Time) error {
	if candidate.IsRunning() {
		if candidate.StartedAt.After(now) {
			return ErrIntervalInFuture
		}
		return nil
	}
	if !candidate.StoppedAt.After(candidate.StartedAt) {
		return ErrEndBeforeStart
	}
	if candidate.StoppedAt.After(now) {
		return ErrIntervalInFuture
	}
	return nil
}

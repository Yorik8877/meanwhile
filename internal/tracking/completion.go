package tracking

import "time"

// Complete takes the task out of the task list and the tray while keeping
// its time in the history. A running task is paused first. Completing a
// completed task keeps its original completion moment.
func (tracker *Tracker) Complete(id TaskID) error {
	return tracker.change(func() error {
		task, found := tracker.tasks[id]
		if !found {
			return ErrTaskNotFound
		}
		if task.IsCompleted() {
			return nil
		}
		now := tracker.now()
		// Paused before the task is marked completed: if marking fails, the
		// task is merely paused, never completed with a running timer.
		if interval, running := tracker.runningInterval(id); running {
			if err := tracker.storeIntervals([]Interval{interval.closedAt(now)}); err != nil {
				return err
			}
		}
		task.CompletedAt = now
		return tracker.storeTask(task)
	})
}

// Reopen returns a completed task to the task list, at its place by
// creation time.
func (tracker *Tracker) Reopen(id TaskID) error {
	return tracker.change(func() error {
		task, found := tracker.tasks[id]
		if !found {
			return ErrTaskNotFound
		}
		task.CompletedAt = time.Time{}
		return tracker.storeTask(task)
	})
}

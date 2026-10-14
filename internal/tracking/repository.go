package tracking

import "time"

// Repository persists tasks, intervals and the crash-recovery heartbeat.
// DeleteTask must also delete the intervals of the task.
type Repository interface {
	Tasks() ([]Task, error)
	Intervals() ([]Interval, error)
	InsertTask(task Task) (TaskID, error)
	UpdateTask(task Task) error
	DeleteTask(id TaskID) error
	InsertInterval(interval Interval) (IntervalID, error)
	UpdateIntervals(intervals []Interval) error
	DeleteInterval(id IntervalID) error
	// Heartbeat returns the zero time when no heartbeat was saved yet.
	Heartbeat() (time.Time, error)
	SaveHeartbeat(moment time.Time) error
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

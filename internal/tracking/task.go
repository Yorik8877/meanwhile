package tracking

import "time"

type TaskID int64

type IntervalID int64

// Task is in progress while CompletedAt is zero.
type Task struct {
	ID          TaskID
	Name        string
	Note        string
	CreatedAt   time.Time
	CompletedAt time.Time
}

func (task Task) IsCompleted() bool {
	return !task.CompletedAt.IsZero()
}

// Interval is a span of work on a task. A zero StoppedAt means the interval
// is still running.
type Interval struct {
	ID        IntervalID
	TaskID    TaskID
	StartedAt time.Time
	StoppedAt time.Time
}

func (interval Interval) IsRunning() bool {
	return interval.StoppedAt.IsZero()
}

// EndAt treats a running interval as ending at now.
func (interval Interval) EndAt(now time.Time) time.Time {
	if interval.IsRunning() {
		return now
	}
	return interval.StoppedAt
}

func (interval Interval) DurationAt(now time.Time) time.Duration {
	return interval.OverlapWith(interval.StartedAt, interval.EndAt(now), now)
}

// OverlapWith returns how much of the interval falls into [rangeStart, rangeEnd).
func (interval Interval) OverlapWith(rangeStart, rangeEnd, now time.Time) time.Duration {
	start := later(interval.StartedAt, rangeStart)
	end := earlier(interval.EndAt(now), rangeEnd)
	if !end.After(start) {
		return 0
	}
	return end.Sub(start)
}

func (interval Interval) overlaps(other Interval, now time.Time) bool {
	return interval.StartedAt.Before(other.EndAt(now)) && other.StartedAt.Before(interval.EndAt(now))
}

// closedAt never closes an interval before its start, so moving the system
// clock backwards cannot produce a negative duration.
func (interval Interval) closedAt(moment time.Time) Interval {
	interval.StoppedAt = later(interval.StartedAt, moment)
	return interval
}

func later(first, second time.Time) time.Time {
	if first.After(second) {
		return first
	}
	return second
}

func earlier(first, second time.Time) time.Time {
	if first.Before(second) {
		return first
	}
	return second
}

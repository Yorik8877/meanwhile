package tracking

import "errors"

var (
	ErrEmptyName          = errors.New("task name is empty")
	ErrTaskNotFound       = errors.New("task not found")
	ErrIntervalNotFound   = errors.New("interval not found")
	ErrMissingEnd         = errors.New("interval end is missing")
	ErrEndBeforeStart     = errors.New("interval end is not after its start")
	ErrIntervalInFuture   = errors.New("interval extends into the future")
	ErrIntervalsOverlap   = errors.New("interval overlaps another interval of the same task")
	ErrRunningIntervalEnd = errors.New("running interval cannot be given an end")
)

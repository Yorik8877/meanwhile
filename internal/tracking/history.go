package tracking

import (
	"cmp"
	"slices"
	"time"
)

type TaskDuration struct {
	Task     Task
	Duration time.Duration
}

type DayTotal struct {
	Date  Date
	Total time.Duration
	Tasks []TaskDuration
}

// DailyTotals returns the work done on each of dayCount days starting at
// first, newest day first. Days without any work are omitted. Within a day
// tasks are ordered by duration, longest first.
func (tracker *Tracker) DailyTotals(first Date, dayCount int) []DayTotal {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	now := tracker.now()
	var totals []DayTotal
	for offset := dayCount - 1; offset >= 0; offset-- {
		dayTotal := tracker.dayTotal(first.AddDays(offset), now)
		if dayTotal.Total == 0 {
			continue
		}
		totals = append(totals, dayTotal)
	}
	return totals
}

func (tracker *Tracker) dayTotal(date Date, now time.Time) DayTotal {
	dayStart := date.Start(tracker.location)
	nextDayStart := date.AddDays(1).Start(tracker.location)

	durations := make(map[TaskID]time.Duration)
	for _, interval := range tracker.intervals {
		durations[interval.TaskID] += interval.OverlapWith(dayStart, nextDayStart, now)
	}

	dayTotal := DayTotal{Date: date}
	for _, task := range tracker.tasksNewestFirst() {
		duration := durations[task.ID]
		if duration == 0 {
			continue
		}
		dayTotal.Tasks = append(dayTotal.Tasks, TaskDuration{Task: task, Duration: duration})
		dayTotal.Total += duration
	}
	slices.SortStableFunc(dayTotal.Tasks, func(first, second TaskDuration) int {
		return cmp.Compare(second.Duration, first.Duration)
	})
	return dayTotal
}

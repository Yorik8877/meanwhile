package tracking

import (
	"cmp"
	"slices"
	"time"
)

type TaskSummary struct {
	Task    Task
	Running bool
	Today   time.Duration
	Total   time.Duration
}

// ActiveTaskSummaries lists the tasks in progress, newest first.
func (tracker *Tracker) ActiveTaskSummaries() []TaskSummary {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	active := slices.DeleteFunc(tracker.tasksNewestFirst(), Task.IsCompleted)
	return tracker.summarize(active)
}

// CompletedTaskSummaries lists the completed tasks, most recently completed
// first.
func (tracker *Tracker) CompletedTaskSummaries() []TaskSummary {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	completed := slices.DeleteFunc(tracker.tasksNewestFirst(), func(task Task) bool { return !task.IsCompleted() })
	slices.SortStableFunc(completed, func(first, second Task) int {
		return second.CompletedAt.Compare(first.CompletedAt)
	})
	return tracker.summarize(completed)
}

// summarize keeps the order of tasks.
func (tracker *Tracker) summarize(tasks []Task) []TaskSummary {
	now := tracker.now()
	today := DateOf(now.In(tracker.location))
	todayStart := today.Start(tracker.location)
	tomorrowStart := today.AddDays(1).Start(tracker.location)

	summaries := make([]TaskSummary, 0, len(tasks))
	indexByTask := make(map[TaskID]int, len(tasks))
	for _, task := range tasks {
		indexByTask[task.ID] = len(summaries)
		summaries = append(summaries, TaskSummary{Task: task})
	}
	for _, interval := range tracker.intervals {
		index, found := indexByTask[interval.TaskID]
		if !found {
			continue
		}
		summary := &summaries[index]
		summary.Total += interval.DurationAt(now)
		summary.Today += interval.OverlapWith(todayStart, tomorrowStart, now)
		summary.Running = summary.Running || interval.IsRunning()
	}
	return summaries
}

// Intervals lists the intervals of a task, newest first.
func (tracker *Tracker) Intervals(taskID TaskID) []Interval {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	var intervals []Interval
	for _, interval := range tracker.intervals {
		if interval.TaskID == taskID {
			intervals = append(intervals, interval)
		}
	}
	slices.SortFunc(intervals, func(first, second Interval) int {
		return second.StartedAt.Compare(first.StartedAt)
	})
	return intervals
}

func (tracker *Tracker) AnyRunning() bool {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	for _, interval := range tracker.intervals {
		if interval.IsRunning() {
			return true
		}
	}
	return false
}

func (tracker *Tracker) tasksNewestFirst() []Task {
	tasks := make([]Task, 0, len(tracker.tasks))
	for _, task := range tracker.tasks {
		tasks = append(tasks, task)
	}
	slices.SortFunc(tasks, func(first, second Task) int {
		if order := second.CreatedAt.Compare(first.CreatedAt); order != 0 {
			return order
		}
		return cmp.Compare(second.ID, first.ID)
	})
	return tasks
}

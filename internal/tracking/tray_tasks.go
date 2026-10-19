package tracking

import (
	"cmp"
	"slices"
	"time"
)

type TrayTask struct {
	Task    Task
	Running bool
}

// TrayTasks returns running tasks first, then the most recently active ones,
// up to limit tasks. Running tasks are never cut off by the limit.
func (tracker *Tracker) TrayTasks(limit int) []TrayTask {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	lastActivity := make(map[TaskID]time.Time, len(tracker.tasks))
	running := make(map[TaskID]bool)
	for _, task := range tracker.tasks {
		lastActivity[task.ID] = task.CreatedAt
	}
	for _, interval := range tracker.intervals {
		if interval.IsRunning() {
			running[interval.TaskID] = true
		}
		lastActivity[interval.TaskID] = later(lastActivity[interval.TaskID], interval.StartedAt)
	}

	trayTasks := make([]TrayTask, 0, len(tracker.tasks))
	for _, task := range tracker.tasks {
		if task.IsCompleted() {
			continue
		}
		trayTasks = append(trayTasks, TrayTask{Task: task, Running: running[task.ID]})
	}
	slices.SortFunc(trayTasks, func(first, second TrayTask) int {
		if first.Running != second.Running {
			if first.Running {
				return -1
			}
			return 1
		}
		if order := lastActivity[second.Task.ID].Compare(lastActivity[first.Task.ID]); order != 0 {
			return order
		}
		return cmp.Compare(second.Task.ID, first.Task.ID)
	})

	visibleCount := min(max(limit, len(running)), len(trayTasks))
	return trayTasks[:visibleCount]
}

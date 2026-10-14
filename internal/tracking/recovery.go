package tracking

import "fmt"

func (tracker *Tracker) RecordHeartbeat() error {
	if err := tracker.repository.SaveHeartbeat(tracker.now()); err != nil {
		return fmt.Errorf("save heartbeat: %w", err)
	}
	return nil
}

// RecoverAfterCrash closes the intervals left running by a previous run that
// did not exit cleanly. They are closed at the last heartbeat, so the time
// after the crash is not counted. Without a heartbeat they get zero length.
func (tracker *Tracker) RecoverAfterCrash() error {
	heartbeat, err := tracker.repository.Heartbeat()
	if err != nil {
		return fmt.Errorf("read heartbeat: %w", err)
	}

	return tracker.change(func() error {
		var closed []Interval
		for _, interval := range tracker.intervals {
			if interval.IsRunning() {
				closed = append(closed, interval.closedAt(heartbeat))
			}
		}
		if len(closed) == 0 {
			return nil
		}
		return tracker.storeIntervals(closed)
	})
}

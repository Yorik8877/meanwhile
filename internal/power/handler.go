// Package power pauses all timers when the computer goes to sleep or shuts
// down, using the login1 D-Bus API of systemd-logind or elogind.
package power

import (
	"log/slog"
	"time"
)

type Pauser interface {
	PauseAll(moment time.Time) error
}

// inhibitor delays sleep and shutdown while it is held, giving the program
// time to record the pause.
type inhibitor interface {
	acquire() error
	release()
}

const (
	prepareForSleepSignal    = managerInterface + ".PrepareForSleep"
	prepareForShutdownSignal = managerInterface + ".PrepareForShutdown"
)

type signalHandler struct {
	pauser    Pauser
	inhibitor inhibitor
	now       func() time.Time
	logger    *slog.Logger
}

// handle reacts to PrepareForSleep and PrepareForShutdown. starting is true
// right before the transition and false after waking up or when a shutdown
// is cancelled.
func (handler *signalHandler) handle(signalName string, starting bool) {
	if signalName != prepareForSleepSignal && signalName != prepareForShutdownSignal {
		return
	}
	if !starting {
		if err := handler.inhibitor.acquire(); err != nil {
			handler.logger.Warn("cannot delay the next sleep", "error", err)
		}
		return
	}

	if err := handler.pauser.PauseAll(handler.now()); err != nil {
		handler.logger.Error("cannot pause timers before sleep or shutdown", "error", err)
	}
	// Released even after a failed pause: holding it would only delay sleep.
	handler.inhibitor.release()
}

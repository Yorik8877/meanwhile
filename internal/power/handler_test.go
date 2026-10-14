package power

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

type recordingPauser struct {
	pausedAt []time.Time
	err      error
}

func (pauser *recordingPauser) PauseAll(moment time.Time) error {
	pauser.pausedAt = append(pauser.pausedAt, moment)
	return pauser.err
}

type countingInhibitor struct {
	acquired int
	released int
}

func (inhibitor *countingInhibitor) acquire() error {
	inhibitor.acquired++
	return nil
}

func (inhibitor *countingInhibitor) release() {
	inhibitor.released++
}

var signalMoment = time.Date(2026, time.October, 14, 18, 0, 0, 0, time.UTC)

func newHandler(pauser *recordingPauser, inhibitor *countingInhibitor) *signalHandler {
	return &signalHandler{
		pauser:    pauser,
		inhibitor: inhibitor,
		now:       func() time.Time { return signalMoment },
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestSleepAndShutdownPauseTimersThenReleaseInhibitor(t *testing.T) {
	for _, signalName := range []string{prepareForSleepSignal, prepareForShutdownSignal} {
		t.Run(signalName, func(t *testing.T) {
			pauser := &recordingPauser{}
			inhibitor := &countingInhibitor{}

			newHandler(pauser, inhibitor).handle(signalName, true)

			if len(pauser.pausedAt) != 1 || !pauser.pausedAt[0].Equal(signalMoment) {
				t.Errorf("paused at %v, want once at %v", pauser.pausedAt, signalMoment)
			}
			if inhibitor.released != 1 {
				t.Errorf("released %d times, want 1", inhibitor.released)
			}
		})
	}
}

func TestWakeUpTakesInhibitorAgainWithoutPausing(t *testing.T) {
	pauser := &recordingPauser{}
	inhibitor := &countingInhibitor{}

	newHandler(pauser, inhibitor).handle(prepareForSleepSignal, false)

	if inhibitor.acquired != 1 || len(pauser.pausedAt) != 0 {
		t.Errorf("acquired %d, paused %d; want 1 and 0", inhibitor.acquired, len(pauser.pausedAt))
	}
}

func TestFailedPauseStillReleasesInhibitor(t *testing.T) {
	pauser := &recordingPauser{err: errors.New("disk full")}
	inhibitor := &countingInhibitor{}

	newHandler(pauser, inhibitor).handle(prepareForSleepSignal, true)

	if inhibitor.released != 1 {
		t.Errorf("released %d times, want 1", inhibitor.released)
	}
}

func TestUnrelatedSignalsAreIgnored(t *testing.T) {
	pauser := &recordingPauser{}
	inhibitor := &countingInhibitor{}

	newHandler(pauser, inhibitor).handle(managerInterface+".SessionNew", true)

	if len(pauser.pausedAt) != 0 || inhibitor.acquired != 0 || inhibitor.released != 0 {
		t.Error("unrelated signal changed state")
	}
}

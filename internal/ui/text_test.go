package ui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

func TestFormatDuration(t *testing.T) {
	cases := map[time.Duration]string{
		0:                "0:00:00",
		59 * time.Second: "0:00:59",
		time.Hour + 5*time.Minute + 32*time.Second: "1:05:32",
		25*time.Hour + 999*time.Millisecond:        "25:00:00",
		-5 * time.Second:                           "0:00:00",
	}
	for duration, want := range cases {
		if got := formatDuration(duration); got != want {
			t.Errorf("formatDuration(%v) = %q, want %q", duration, got, want)
		}
	}
}

func TestComposeMoment(t *testing.T) {
	location := time.UTC
	date := time.Date(2026, time.October, 14, 0, 0, 0, 0, location)
	original := time.Date(2026, time.October, 14, 9, 30, 42, 0, location)

	unchanged, err := composeMoment(&date, "09:30", original, location)
	if err != nil || !unchanged.Equal(original) {
		t.Errorf("unchanged field: got %v, %v; want original with seconds", unchanged, err)
	}

	changed, err := composeMoment(&date, " 9:45 ", original, location)
	want := time.Date(2026, time.October, 14, 9, 45, 0, 0, location)
	if err != nil || !changed.Equal(want) {
		t.Errorf("changed field: got %v, %v; want %v", changed, err, want)
	}

	for _, badText := range []string{"", "24:00", "9:7", "09-30", "abc"} {
		if _, err := composeMoment(&date, badText, time.Time{}, location); !errors.Is(err, errBadClock) {
			t.Errorf("text %q: error = %v, want errBadClock", badText, err)
		}
	}

	if _, err := composeMoment(nil, "09:30", time.Time{}, location); !errors.Is(err, errMissingDate) {
		t.Errorf("missing date: error = %v, want errMissingDate", err)
	}
}

// userFacingErrors are the errors a person can cause from the interface;
// each of them needs its own translated sentence.
var userFacingErrors = []error{
	tracking.ErrEmptyName,
	tracking.ErrTaskNotFound,
	tracking.ErrIntervalNotFound,
	tracking.ErrMissingEnd,
	tracking.ErrEndBeforeStart,
	tracking.ErrIntervalInFuture,
	tracking.ErrIntervalsOverlap,
	tracking.ErrRunningIntervalEnd,
	errBadClock,
	errMissingDate,
}

func TestUserFacingErrorsAreNotShownAsTechnicalText(t *testing.T) {
	for _, target := range userFacingErrors {
		wrapped := fmt.Errorf("context: %w", target)
		if text := errorText(wrapped); strings.Contains(text, target.Error()) {
			t.Errorf("error %q is shown as technical text %q", target, text)
		}
	}
}

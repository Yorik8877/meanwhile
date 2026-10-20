package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func TestStartButtonShowsPauseWhileRunning(t *testing.T) {
	test.NewTempApp(t)
	button := newStartButton()

	button.SetRunning(true)

	pauseOnPrimary := theme.NewColoredResource(theme.MediaPauseIcon(), theme.ColorNameForegroundOnPrimary)
	if button.icon.Resource.Name() != pauseOnPrimary.Name() {
		t.Errorf("icon = %s, want %s", button.icon.Resource.Name(), pauseOnPrimary.Name())
	}
	if button.tapArea.Importance != widget.HighImportance {
		t.Errorf("importance = %v, want high while running", button.tapArea.Importance)
	}

	button.SetRunning(false)

	if button.icon.Resource.Name() != theme.MediaPlayIcon().Name() {
		t.Errorf("icon = %s, want the play icon", button.icon.Resource.Name())
	}
	if button.tapArea.Importance != widget.LowImportance {
		t.Errorf("importance = %v, want low while paused", button.tapArea.Importance)
	}
}

func TestStartButtonCallsOnTapped(t *testing.T) {
	test.NewTempApp(t)
	button := newStartButton()
	taps := 0
	button.OnTapped = func() { taps++ }

	test.Tap(button.tapArea)

	if taps != 1 {
		t.Errorf("OnTapped ran %d times, want once", taps)
	}
}

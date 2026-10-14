package ui

import (
	"errors"
	"testing"

	"fyne.io/fyne/v2/test"
)

type fakeAutostart struct {
	enabled   bool
	saveError error
}

func (autostart *fakeAutostart) Enabled() (bool, error) {
	return autostart.enabled, nil
}

func (autostart *fakeAutostart) SetEnabled(enabled bool) error {
	if autostart.saveError != nil {
		return autostart.saveError
	}
	autostart.enabled = enabled
	return nil
}

func newSettingsTestWindow(t *testing.T, autostart *fakeAutostart) *MainWindow {
	t.Helper()
	test.NewTempApp(t)
	return &MainWindow{window: test.NewTempWindow(t, nil), settings: Settings{Autostart: autostart}}
}

func TestAutostartCheckShowsAndSavesState(t *testing.T) {
	autostart := &fakeAutostart{enabled: true}
	check := newSettingsTestWindow(t, autostart).autostartCheck()

	if !check.Checked {
		t.Fatal("check is off although autostart is enabled")
	}
	test.Tap(check)

	if autostart.enabled {
		t.Error("autostart stays enabled after unchecking")
	}
}

func TestAutostartCheckRevertsWhenSaveFails(t *testing.T) {
	autostart := &fakeAutostart{saveError: errors.New("read-only file system")}
	check := newSettingsTestWindow(t, autostart).autostartCheck()

	test.Tap(check)

	if check.Checked {
		t.Error("check shows autostart enabled although saving failed")
	}
}

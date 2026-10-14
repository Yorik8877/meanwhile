package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

// Fyne remembers each tray item's action before it fills in the action of
// IsQuit items, so a quit item rebuilt without an action never quits.
func TestTrayQuitItemHasItsOwnAction(t *testing.T) {
	test.NewTempApp(t)
	tray := &tray{tracker: newTestTracker(t), window: test.NewTempWindow(t, nil)}

	items := tray.menuItems()

	quit := items[len(items)-1]
	if !quit.IsQuit || quit.Action == nil {
		t.Errorf("last tray item %q: IsQuit = %v, has action = %v; want a quit item with an action",
			quit.Label, quit.IsQuit, quit.Action != nil)
	}
}

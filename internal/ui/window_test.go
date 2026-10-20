package ui

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

// Fyne 2.8 gives a theme override a new font cache scope on every refresh
// and never frees the old ones. Switching tabs refreshes the content of
// every tab, so any override inside the window leaks memory.
func TestWindowContentHasNoThemeOverride(t *testing.T) {
	test.NewTempApp(t)
	tracker := newTestTracker(t)
	runningID, err := tracker.CreateTask("Running", "Note")
	if err != nil {
		t.Fatal(err)
	}
	if err := tracker.Start(runningID); err != nil {
		t.Fatal(err)
	}
	completedID, err := tracker.CreateTask("Completed", "Note")
	if err != nil {
		t.Fatal(err)
	}
	if err := tracker.Complete(completedID); err != nil {
		t.Fatal(err)
	}
	mainWindow := &MainWindow{window: test.NewTempWindow(t, nil), tracker: tracker}
	mainWindow.build()
	mainWindow.window.Resize(windowSize)

	for _, item := range mainWindow.tabs.Items {
		mainWindow.tabs.Select(item)
		if path := themeOverridePath(item.Content); path != "" {
			t.Errorf("tab %q holds a theme override: %s", item.Text, path)
		}
	}
}

// themeOverridePath describes where the first theme override under object
// is, or returns "" when there is none.
func themeOverridePath(object fyne.CanvasObject) string {
	if _, isOverride := object.(*container.ThemeOverride); isOverride {
		return "ThemeOverride"
	}
	for _, child := range childObjects(object) {
		if path := themeOverridePath(child); path != "" {
			return typeName(object) + " > " + path
		}
	}
	return ""
}

func childObjects(object fyne.CanvasObject) []fyne.CanvasObject {
	switch typed := object.(type) {
	case *fyne.Container:
		return typed.Objects
	case fyne.Widget:
		return test.WidgetRenderer(typed).Objects()
	default:
		return nil
	}
}

func typeName(object fyne.CanvasObject) string {
	return fmt.Sprintf("%T", object)
}

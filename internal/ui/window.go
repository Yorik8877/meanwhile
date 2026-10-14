package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"stopwatch/internal/tracking"
)

type MainWindow struct {
	window    fyne.Window
	tabs      *container.AppTabs
	tasksItem *container.TabItem
	tasks     *tasksTab
	history   *historyTab
}

// NewMainWindow builds the window and the tray icon. Closing the window only
// hides it; the program keeps running in the tray.
func NewMainWindow(application fyne.App, tracker *tracking.Tracker) *MainWindow {
	window := application.NewWindow(message{"app.title", "Stopwatch"}.String())
	mainWindow := &MainWindow{
		window:  window,
		tasks:   newTasksTab(tracker, window),
		history: newHistoryTab(tracker),
	}
	mainWindow.tasksItem = container.NewTabItem(message{"tab.tasks", "Tasks"}.String(), mainWindow.tasks.content)
	historyItem := container.NewTabItem(message{"tab.history", "History"}.String(), mainWindow.history.content)
	mainWindow.tabs = container.NewAppTabs(mainWindow.tasksItem, historyItem)
	mainWindow.tabs.OnSelected = func(*container.TabItem) { mainWindow.refresh() }

	window.SetContent(mainWindow.tabs)
	window.Resize(windowSize)
	window.SetCloseIntercept(window.Hide)
	tracker.Subscribe(func() { fyne.Do(mainWindow.refresh) })
	installTray(application, window, tracker)
	return mainWindow
}

func (mainWindow *MainWindow) Show() {
	mainWindow.window.Show()
	mainWindow.window.RequestFocus()
}

// RunTicker redraws the running timers every second until stop is closed.
func (mainWindow *MainWindow) RunTicker(stop <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			fyne.Do(mainWindow.refresh)
		}
	}
}

func (mainWindow *MainWindow) refresh() {
	if mainWindow.tabs.Selected() == mainWindow.tasksItem {
		mainWindow.tasks.refresh()
		return
	}
	mainWindow.history.refresh()
}

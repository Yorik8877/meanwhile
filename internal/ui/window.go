package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

type MainWindow struct {
	window        fyne.Window
	tracker       *tracking.Tracker
	settings      Settings
	tray          *tray
	tabs          *container.AppTabs
	tasksItem     *container.TabItem
	completedItem *container.TabItem
	tasks         *tasksTab
	completed     *completedTab
	history       *historyTab
}

// NewMainWindow builds the window and the tray icon. Closing the window only
// hides it; the program keeps running in the tray.
func NewMainWindow(application fyne.App, tracker *tracking.Tracker, settings Settings) *MainWindow {
	mainWindow := &MainWindow{
		window:   application.NewWindow(""),
		tracker:  tracker,
		settings: settings,
	}
	mainWindow.build()
	mainWindow.window.Resize(windowSize)
	mainWindow.window.SetCloseIntercept(mainWindow.window.Hide)
	tracker.Subscribe(func() { fyne.Do(mainWindow.refresh) })
	mainWindow.tray = installTray(application, mainWindow.window, tracker)
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
			fyne.Do(mainWindow.tick)
		}
	}
}

// tick redraws only what changes with time alone, on the visible tab.
func (mainWindow *MainWindow) tick() {
	switch mainWindow.tabs.Selected() {
	case mainWindow.tasksItem:
		mainWindow.tasks.tick()
	case mainWindow.completedItem:
		// Completed tasks do not run, so their times stand still.
	default:
		mainWindow.history.refresh()
	}
}

// build creates the window content in the current language. It runs again
// after the language changes, because every widget takes its text when it
// is created.
func (mainWindow *MainWindow) build() {
	mainWindow.window.SetTitle(message{"app.title", "Meanwhile"}.String())
	mainWindow.tasks = newTasksTab(mainWindow.tracker, mainWindow.window)
	mainWindow.completed = newCompletedTab(mainWindow.tracker, mainWindow.window)
	mainWindow.history = newHistoryTab(mainWindow.tracker)
	mainWindow.tasksItem = container.NewTabItem(message{"tab.tasks", "Tasks"}.String(), mainWindow.tasks.content)
	mainWindow.completedItem = container.NewTabItem(message{"tab.completed", "Completed"}.String(), mainWindow.completed.content)
	historyItem := container.NewTabItem(message{"tab.history", "History"}.String(), mainWindow.history.content)
	mainWindow.tabs = container.NewAppTabs(mainWindow.tasksItem, mainWindow.completedItem, historyItem)
	mainWindow.tabs.OnSelected = func(*container.TabItem) { mainWindow.refresh() }

	// AppTabs has no place for extra buttons, so the settings button is laid
	// over the empty right end of the tab bar.
	var settingsButton *widget.Button
	settingsButton = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		mainWindow.showSettings(settingsButton)
	})
	settingsButton.Importance = widget.LowImportance
	settingsBar := container.NewVBox(container.NewHBox(layout.NewSpacer(), settingsButton))
	mainWindow.window.SetContent(container.NewStack(mainWindow.tabs, settingsBar))
}

func (mainWindow *MainWindow) switchLanguage(code string) {
	if err := mainWindow.settings.Language.SaveLanguage(code); err != nil {
		reportError(mainWindow.window, err)
		return
	}
	if err := ApplyLanguage(code); err != nil {
		reportError(mainWindow.window, err)
		return
	}
	mainWindow.build()
	if mainWindow.tray != nil {
		mainWindow.tray.refresh()
	}
}

// refresh redraws only the visible tab; the others refresh when selected.
func (mainWindow *MainWindow) refresh() {
	switch mainWindow.tabs.Selected() {
	case mainWindow.tasksItem:
		mainWindow.tasks.refresh()
	case mainWindow.completedItem:
		mainWindow.completed.refresh()
	default:
		mainWindow.history.refresh()
	}
}

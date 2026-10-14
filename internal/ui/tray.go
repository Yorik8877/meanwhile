package ui

import (
	"log/slog"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	"stopwatch/internal/tracking"
)

const trayTaskLimit = 10

type tray struct {
	desktopApp       desktop.App
	tracker          *tracking.Tracker
	window           fyne.Window
	iconSet          bool
	showsRunningIcon bool
}

func installTray(application fyne.App, window fyne.Window, tracker *tracking.Tracker) {
	desktopApp, supported := application.(desktop.App)
	if !supported {
		slog.Warn("system tray is not supported by this driver")
		return
	}
	tray := &tray{desktopApp: desktopApp, tracker: tracker, window: window}
	tray.refresh()
	desktopApp.SetSystemTrayWindow(window)
	tracker.Subscribe(func() { fyne.Do(tray.refresh) })
}

func (tray *tray) refresh() {
	tray.desktopApp.SetSystemTrayMenu(fyne.NewMenu(message{"app.title", "Stopwatch"}.String(), tray.menuItems()...))
	tray.updateIcon(tray.tracker.AnyRunning())
}

func (tray *tray) menuItems() []*fyne.MenuItem {
	items := []*fyne.MenuItem{fyne.NewMenuItem(message{"tray.open", "Open"}.String(), tray.showWindow)}

	trayTasks := tray.tracker.TrayTasks(trayTaskLimit)
	if len(trayTasks) > 0 {
		items = append(items, fyne.NewMenuItemSeparator())
	}
	for _, trayTask := range trayTasks {
		taskID := trayTask.Task.ID
		item := fyne.NewMenuItem(trayTask.Task.Name, func() { tray.report(tray.tracker.Toggle(taskID)) })
		item.Checked = trayTask.Running
		items = append(items, item)
	}

	pauseAll := fyne.NewMenuItem(message{"tray.pause_all", "Pause all"}.String(), func() {
		tray.report(tray.tracker.PauseAll(time.Now()))
	})
	pauseAll.Disabled = !tray.tracker.AnyRunning()
	// An IsQuit item without an action makes Fyne quit the application.
	quit := fyne.NewMenuItem(message{"tray.quit", "Quit"}.String(), nil)
	quit.IsQuit = true
	return append(items, fyne.NewMenuItemSeparator(), pauseAll, quit)
}

func (tray *tray) updateIcon(running bool) {
	if tray.iconSet && tray.showsRunningIcon == running {
		return
	}
	icon := AppIcon
	if running {
		icon = runningIcon
	}
	tray.desktopApp.SetSystemTrayIcon(icon)
	tray.iconSet = true
	tray.showsRunningIcon = running
}

func (tray *tray) showWindow() {
	tray.window.Show()
	tray.window.RequestFocus()
}

// report shows the window first, because the error dialog lives in it.
func (tray *tray) report(err error) {
	if err == nil {
		return
	}
	tray.showWindow()
	reportError(tray.window, err)
}

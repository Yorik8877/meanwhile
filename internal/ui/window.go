package ui

import (
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"stopwatch/internal/tracking"
)

type MainWindow struct {
	window    fyne.Window
	tracker   *tracking.Tracker
	languages LanguageSetting
	tray      *tray
	tabs      *container.AppTabs
	tasksItem *container.TabItem
	tasks     *tasksTab
	history   *historyTab
}

// NewMainWindow builds the window and the tray icon. Closing the window only
// hides it; the program keeps running in the tray.
func NewMainWindow(application fyne.App, tracker *tracking.Tracker, languages LanguageSetting) *MainWindow {
	mainWindow := &MainWindow{
		window:    application.NewWindow(""),
		tracker:   tracker,
		languages: languages,
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
			fyne.Do(mainWindow.refresh)
		}
	}
}

// build creates the window content in the current language. It runs again
// after the language changes, because every widget takes its text when it
// is created.
func (mainWindow *MainWindow) build() {
	mainWindow.window.SetTitle(message{"app.title", "Stopwatch"}.String())
	mainWindow.tasks = newTasksTab(mainWindow.tracker, mainWindow.window)
	mainWindow.history = newHistoryTab(mainWindow.tracker)
	mainWindow.tasksItem = container.NewTabItem(message{"tab.tasks", "Tasks"}.String(), mainWindow.tasks.content)
	historyItem := container.NewTabItem(message{"tab.history", "History"}.String(), mainWindow.history.content)
	mainWindow.tabs = container.NewAppTabs(mainWindow.tasksItem, historyItem)
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

// showSettings opens a small panel under the settings button, kept away
// from the window's right edge.
func (mainWindow *MainWindow) showSettings(anchor fyne.CanvasObject) {
	current, err := mainWindow.languages.Language()
	if err != nil {
		reportError(mainWindow.window, err)
		return
	}

	labels := make([]string, len(languageOptions))
	for index, option := range languageOptions {
		labels[index] = option.label()
	}
	selector := widget.NewSelect(labels, nil)
	selector.SetSelected(languageOptionFor(current).label())
	content := container.NewHBox(widget.NewLabel(message{"settings.language", "Language:"}.String()), selector)

	windowCanvas := mainWindow.window.Canvas()
	panel := widget.NewPopUp(content, windowCanvas)
	// Assigned after SetSelected, so that showing the current value does not
	// count as a change.
	selector.OnChanged = func(label string) {
		panel.Hide()
		mainWindow.switchLanguage(languageOptions[slices.Index(labels, label)].code)
	}

	anchorPosition := fyne.CurrentApp().Driver().AbsolutePositionForObject(anchor)
	panelLeft := windowCanvas.Size().Width - panel.MinSize().Width - settingsPanelMargin
	panel.ShowAtPosition(fyne.NewPos(panelLeft, anchorPosition.Y+anchor.Size().Height))
}

func (mainWindow *MainWindow) switchLanguage(code string) {
	if err := mainWindow.languages.SaveLanguage(code); err != nil {
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

func (mainWindow *MainWindow) refresh() {
	if mainWindow.tabs.Selected() == mainWindow.tasksItem {
		mainWindow.tasks.refresh()
		return
	}
	mainWindow.history.refresh()
}

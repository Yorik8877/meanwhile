// Command stopwatch tracks time spent on tasks that run in parallel. It lives
// in the system tray; a left click on the tray icon opens its window.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"stopwatch/internal/instance"
	"stopwatch/internal/power"
	"stopwatch/internal/storage"
	"stopwatch/internal/tracking"
	"stopwatch/internal/ui"
)

const (
	applicationID     = "io.github.stopwatch"
	databaseFileName  = "stopwatch.db"
	heartbeatInterval = time.Minute
	exitSuccess       = 0
	exitFailure       = 1
)

type options struct {
	hidden          bool
	dataDirOverride string
}

func main() {
	os.Exit(run(parseOptions()))
}

func parseOptions() options {
	var parsed options
	flag.BoolVar(&parsed.hidden, "hidden", false, "start in the tray without showing the window")
	flag.StringVar(&parsed.dataDirOverride, "data-dir", "", "keep the database, the log and the instance lock in this directory")
	flag.Parse()
	return parsed
}

func run(options options) int {
	resolved, err := prepareLocations(options.dataDirOverride)
	if err != nil {
		fmt.Fprintln(os.Stderr, "stopwatch:", err)
		return exitFailure
	}
	closeLog, err := setupLogging(resolved.StateDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "stopwatch:", err)
		return exitFailure
	}
	defer closeLog()

	guard, err := instance.Acquire(resolved.RuntimeDir)
	if errors.Is(err, instance.ErrAlreadyRunning) {
		return showRunningInstance(resolved.RuntimeDir)
	}
	if err != nil {
		slog.Error("running without single instance protection", "error", err)
	} else {
		defer guard.Close()
	}

	return runApplication(resolved, guard, options.hidden)
}

func showRunningInstance(runtimeDir string) int {
	if err := instance.RequestShow(runtimeDir); err != nil {
		slog.Error("another instance is running but did not answer", "error", err)
		return exitFailure
	}
	return exitSuccess
}

// runApplication blocks until the user quits. guard is nil when single
// instance protection is unavailable.
func runApplication(resolved locations, guard *instance.Guard, hidden bool) int {
	application := app.NewWithID(applicationID)
	application.SetIcon(ui.AppIcon)
	// The system language comes first, so that a database error is shown in
	// it; the saved language needs the database.
	if err := ui.LoadTranslations(); err != nil {
		slog.Error("load translations", "error", err)
	}

	databasePath := filepath.Join(resolved.DataDir, databaseFileName)
	store, tracker, err := openTracker(databasePath)
	if err != nil {
		slog.Error("cannot start", "error", err)
		ui.ShowStartupError(application, databasePath, err)
		return exitFailure
	}
	defer store.Close()
	applySavedLanguage(store)

	mainWindow := ui.NewMainWindow(application, tracker, store)
	if guard != nil {
		go guard.Serve(func() { fyne.Do(mainWindow.Show) })
	}
	stopWatchingPower := watchPower(tracker)
	defer stopWatchingPower()

	stop := make(chan struct{})
	go mainWindow.RunTicker(stop)
	go recordHeartbeats(tracker, stop)
	if !hidden {
		mainWindow.Show()
	}

	application.Run()
	close(stop)
	if err := tracker.PauseAll(time.Now()); err != nil {
		slog.Error("pause timers on exit", "error", err)
	}
	return exitSuccess
}

// openTracker loads the data and closes intervals that a crashed previous run
// left open.
func openTracker(databasePath string) (*storage.Store, *tracking.Tracker, error) {
	store, err := storage.Open(databasePath)
	if err != nil {
		return nil, nil, err
	}
	tracker := tracking.NewTracker(store, tracking.SystemClock{}, time.Local)
	if err := tracker.Load(); err != nil {
		store.Close()
		return nil, nil, err
	}
	if err := tracker.RecoverAfterCrash(); err != nil {
		slog.Error("recover intervals after crash", "error", err)
	}
	return store, tracker, nil
}

// watchPower returns a function that stops watching.
func watchPower(tracker *tracking.Tracker) func() {
	watcher, err := power.Watch(tracker, slog.Default())
	if err != nil {
		slog.Warn("timers will not pause during sleep", "error", err)
		return func() {}
	}
	return func() { watcher.Close() }
}

func recordHeartbeats(tracker *tracking.Tracker, stop <-chan struct{}) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		recordHeartbeat(tracker)
		select {
		case <-stop:
			return
		case <-ticker.C:
		}
	}
}

func recordHeartbeat(tracker *tracking.Tracker) {
	if err := tracker.RecordHeartbeat(); err != nil {
		slog.Error("record heartbeat", "error", err)
	}
}

func applySavedLanguage(store *storage.Store) {
	code, err := store.Language()
	if err != nil {
		slog.Error("read saved language", "error", err)
		return
	}
	if err := ui.ApplyLanguage(code); err != nil {
		slog.Error("apply saved language", "error", err)
	}
}

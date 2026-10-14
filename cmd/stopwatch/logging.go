package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	logFileName = "stopwatch.log"
	// maxLogSizeBytes keeps the log from growing forever: a larger log is
	// cleared at startup.
	maxLogSizeBytes = 1 << 20
)

// setupLogging writes logs to stderr and to a file in the state directory.
// The returned function closes the file.
func setupLogging(stateDir string) (func(), error) {
	path := filepath.Join(stateDir, logFileName)
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if info, err := os.Stat(path); err == nil && info.Size() > maxLogSizeBytes {
		flags |= os.O_TRUNC
	}
	file, err := os.OpenFile(path, flags, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open log %s: %w", path, err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(os.Stderr, file), nil)))
	return func() { file.Close() }, nil
}

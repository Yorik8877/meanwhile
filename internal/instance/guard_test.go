package instance_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"stopwatch/internal/instance"
)

func acquire(t *testing.T, directory string) *instance.Guard {
	t.Helper()
	guard, err := instance.Acquire(directory)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	t.Cleanup(func() { guard.Close() })
	return guard
}

func TestSecondAcquireReportsAlreadyRunning(t *testing.T) {
	directory := t.TempDir()
	acquire(t, directory)

	_, err := instance.Acquire(directory)

	if !errors.Is(err, instance.ErrAlreadyRunning) {
		t.Errorf("error = %v, want ErrAlreadyRunning", err)
	}
}

func TestShowRequestReachesRunningInstance(t *testing.T) {
	directory := t.TempDir()
	guard := acquire(t, directory)
	shown := make(chan struct{}, 1)
	go guard.Serve(func() { shown <- struct{}{} })

	if err := instance.RequestShow(directory); err != nil {
		t.Fatalf("request show: %v", err)
	}

	select {
	case <-shown:
	case <-time.After(2 * time.Second):
		t.Fatal("running instance did not receive the show request")
	}
}

func TestLockIsReleasedOnClose(t *testing.T) {
	directory := t.TempDir()
	guard, err := instance.Acquire(directory)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}

	if err := guard.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	acquire(t, directory)
}

func TestStaleSocketFileIsReplaced(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "stopwatch.sock"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	acquire(t, directory)
}

func TestRequestShowFailsWithoutRunningInstance(t *testing.T) {
	if err := instance.RequestShow(t.TempDir()); err == nil {
		t.Error("request show succeeded without a running instance")
	}
}

func TestLongDirectoryPathStillWorks(t *testing.T) {
	directory := filepath.Join(t.TempDir(), strings.Repeat("long-directory-name-", 6))
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	guard := acquire(t, directory)
	shown := make(chan struct{}, 1)
	go guard.Serve(func() { shown <- struct{}{} })

	if err := instance.RequestShow(directory); err != nil {
		t.Fatalf("request show: %v", err)
	}

	select {
	case <-shown:
	case <-time.After(2 * time.Second):
		t.Fatal("running instance did not receive the show request")
	}
}

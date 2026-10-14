// Package instance keeps a single running copy of the program per user. The
// running copy holds an flock on a lock file and listens on a Unix socket;
// a second copy uses the socket to ask the first one to show its window.
package instance

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var ErrAlreadyRunning = errors.New("another instance is already running")

const (
	lockFileName      = "stopwatch.lock"
	socketFileName    = "stopwatch.sock"
	showCommand       = "show"
	connectionTimeout = 2 * time.Second
	// maxSocketPathBytes is the size of sun_path in sockaddr_un on Linux
	// without the terminating zero byte.
	maxSocketPathBytes = 107
	abstractPrefix     = "@"
)

type Guard struct {
	lockFile *os.File
	listener net.Listener
}

// Acquire makes this process the running instance. The kernel releases the
// lock when the process exits, even after a crash, so a lock is never left
// behind.
func Acquire(directory string) (*Guard, error) {
	lockFile, err := os.OpenFile(filepath.Join(directory, lockFileName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lockFile.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrAlreadyRunning
		}
		return nil, fmt.Errorf("lock %s: %w", lockFile.Name(), err)
	}

	listener, err := listen(socketAddress(directory))
	if err != nil {
		lockFile.Close()
		return nil, err
	}
	return &Guard{lockFile: lockFile, listener: listener}, nil
}

// Serve calls onShow for every show request until Close is called. It
// blocks, so run it on its own goroutine.
func (guard *Guard) Serve(onShow func()) {
	for {
		connection, err := guard.listener.Accept()
		if err != nil {
			return
		}
		handleRequest(connection, onShow)
	}
}

func (guard *Guard) Close() error {
	return errors.Join(guard.listener.Close(), guard.lockFile.Close())
}

// RequestShow asks the running instance to show its window.
func RequestShow(directory string) error {
	connection, err := net.DialTimeout("unix", socketAddress(directory), connectionTimeout)
	if err != nil {
		return fmt.Errorf("connect to running instance: %w", err)
	}
	defer connection.Close()

	if _, err := fmt.Fprintln(connection, showCommand); err != nil {
		return fmt.Errorf("send show request: %w", err)
	}
	return nil
}

// socketAddress returns the socket file path, or, when that path does not
// fit into sun_path, a Linux abstract socket name derived from it.
func socketAddress(directory string) string {
	path := filepath.Join(directory, socketFileName)
	if len(path) <= maxSocketPathBytes {
		return path
	}
	digest := sha256.Sum256([]byte(path))
	return abstractPrefix + socketFileName + "-" + hex.EncodeToString(digest[:16])
}

// listen removes a leftover socket file first: holding the lock proves that
// no other instance is using it. Abstract sockets have no file.
func listen(address string) (net.Listener, error) {
	if !strings.HasPrefix(address, abstractPrefix) {
		if err := os.Remove(address); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("remove stale socket: %w", err)
		}
	}
	listener, err := net.Listen("unix", address)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", address, err)
	}
	return listener, nil
}

func handleRequest(connection net.Conn, onShow func()) {
	defer connection.Close()

	connection.SetReadDeadline(time.Now().Add(connectionTimeout))
	line, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		return
	}
	if strings.TrimSpace(line) == showCommand {
		onShow()
	}
}

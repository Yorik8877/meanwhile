package power

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	login1Name       = "org.freedesktop.login1"
	login1Path       = dbus.ObjectPath("/org/freedesktop/login1")
	managerInterface = "org.freedesktop.login1.Manager"
)

type Watcher struct {
	connection *dbus.Conn
	inhibitor  *login1Inhibitor
}

// Watch fails when there is no system bus or no login1 service on it; the
// program then keeps working, but timers are not paused during sleep.
func Watch(pauser Pauser, logger *slog.Logger) (*Watcher, error) {
	connection, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("connect to system bus: %w", err)
	}
	inhibitor := &login1Inhibitor{manager: connection.Object(login1Name, login1Path)}
	if err := inhibitor.acquire(); err != nil {
		connection.Close()
		return nil, fmt.Errorf("take login1 inhibitor lock: %w", err)
	}
	if err := subscribe(connection); err != nil {
		inhibitor.release()
		connection.Close()
		return nil, err
	}

	signals := make(chan *dbus.Signal, 8)
	connection.Signal(signals)
	handler := &signalHandler{pauser: pauser, inhibitor: inhibitor, now: time.Now, logger: logger}
	go dispatchSignals(signals, handler)

	return &Watcher{connection: connection, inhibitor: inhibitor}, nil
}

func (watcher *Watcher) Close() error {
	watcher.inhibitor.release()
	return watcher.connection.Close()
}

func subscribe(connection *dbus.Conn) error {
	for _, member := range []string{"PrepareForSleep", "PrepareForShutdown"} {
		err := connection.AddMatchSignal(
			dbus.WithMatchObjectPath(login1Path),
			dbus.WithMatchInterface(managerInterface),
			dbus.WithMatchMember(member),
		)
		if err != nil {
			return fmt.Errorf("subscribe to %s: %w", member, err)
		}
	}
	return nil
}

// dispatchSignals ends when the connection is closed, because godbus closes
// the channel then.
func dispatchSignals(signals <-chan *dbus.Signal, handler *signalHandler) {
	for signal := range signals {
		if len(signal.Body) != 1 {
			continue
		}
		starting, isBool := signal.Body[0].(bool)
		if !isBool {
			continue
		}
		handler.handle(signal.Name, starting)
	}
}

// login1Inhibitor holds a "delay" inhibitor lock. login1 waits until the
// lock's file descriptor is closed, or until its InhibitDelayMaxSec timeout.
type login1Inhibitor struct {
	manager dbus.BusObject
	mutex   sync.Mutex
	lock    *os.File
}

func (inhibitor *login1Inhibitor) acquire() error {
	inhibitor.mutex.Lock()
	defer inhibitor.mutex.Unlock()

	if inhibitor.lock != nil {
		return nil
	}
	var descriptor dbus.UnixFD
	err := inhibitor.manager.Call(managerInterface+".Inhibit", 0,
		"sleep:shutdown", "Meanwhile", "Pausing running timers", "delay",
	).Store(&descriptor)
	if err != nil {
		return err
	}
	inhibitor.lock = os.NewFile(uintptr(descriptor), "login1-inhibitor")
	return nil
}

func (inhibitor *login1Inhibitor) release() {
	inhibitor.mutex.Lock()
	defer inhibitor.mutex.Unlock()

	if inhibitor.lock == nil {
		return
	}
	inhibitor.lock.Close()
	inhibitor.lock = nil
}

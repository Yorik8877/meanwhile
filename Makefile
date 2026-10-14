# Installs into the user's home directory; no root access is needed.

PREFIX      ?= $(HOME)/.local
BINDIR      := $(PREFIX)/bin
DATA_HOME   := $(or $(XDG_DATA_HOME),$(HOME)/.local/share)
STATE_HOME  := $(or $(XDG_STATE_HOME),$(HOME)/.local/state)
CONFIG_HOME := $(or $(XDG_CONFIG_HOME),$(HOME)/.config)

BINARY         := $(BINDIR)/stopwatch
DESKTOP_FILE   := $(DATA_HOME)/applications/io.github.stopwatch.desktop
ICON_FILE      := $(DATA_HOME)/icons/hicolor/scalable/apps/io.github.stopwatch.svg
AUTOSTART_FILE := $(CONFIG_HOME)/autostart/io.github.stopwatch.desktop
DATA_DIR       := $(DATA_HOME)/stopwatch
LOG_DIR        := $(STATE_HOME)/stopwatch
LOCK_FILE      := $(or $(XDG_RUNTIME_DIR),$(LOG_DIR))/stopwatch.lock

.PHONY: build test install uninstall check-not-running

build:
	go build -ldflags "-s -w" -o bin/stopwatch ./cmd/stopwatch

test:
	go test ./...

install: build
	install -Dm755 bin/stopwatch "$(BINARY)"
	install -Dm644 internal/ui/icons/idle.svg "$(ICON_FILE)"
	install -d "$(dir $(DESKTOP_FILE))" "$(dir $(AUTOSTART_FILE))"
	sed 's|@BINARY@|$(BINARY)|' packaging/stopwatch.desktop.in > "$(DESKTOP_FILE)"
	sed 's|@BINARY@|$(BINARY)|' packaging/stopwatch-autostart.desktop.in > "$(AUTOSTART_FILE)"
	@echo "Installed. Start Stopwatch from the applications menu; it will also start at login."

# The running program holds an flock on LOCK_FILE.
check-not-running:
	@if [ -e "$(LOCK_FILE)" ] && command -v flock >/dev/null && ! flock -n "$(LOCK_FILE)" true; then \
		echo "Stopwatch is running. Quit it from the tray menu, then run make uninstall again."; \
		exit 1; \
	fi

# PURGE=1 also deletes the database and logs after a y/N question. Without a
# terminal the answer is "no".
uninstall: check-not-running
	rm -f "$(BINARY)" "$(DESKTOP_FILE)" "$(ICON_FILE)" "$(AUTOSTART_FILE)"
	@if [ "$(PURGE)" = "1" ]; then \
		printf 'Delete the database and logs in %s and %s? [y/N] ' "$(DATA_DIR)" "$(LOG_DIR)"; \
		answer=""; { read answer < /dev/tty; } 2>/dev/null || true; \
		if [ "$$answer" = "y" ] || [ "$$answer" = "Y" ]; then \
			rm -rf "$(DATA_DIR)" "$(LOG_DIR)"; echo "Data deleted."; \
		else \
			echo "Data kept."; \
		fi; \
	else \
		echo "Data kept in $(DATA_DIR). Use make uninstall PURGE=1 to delete it."; \
	fi

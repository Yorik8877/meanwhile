package autostart_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yorik8877/meanwhile/internal/autostart"
)

func mustSucceed(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func isEnabled(t *testing.T, entry *autostart.Entry) bool {
	t.Helper()
	enabled, err := entry.Enabled()
	mustSucceed(t, err)
	return enabled
}

func readEntry(t *testing.T, configDir string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(configDir, "autostart", "io.github.yorik8877.meanwhile.desktop"))
	mustSucceed(t, err)
	return string(content)
}

func TestEntryIsDisabledWhenFileIsMissing(t *testing.T) {
	entry := autostart.NewEntry(t.TempDir(), "/home/user/.local/bin/meanwhile")

	if isEnabled(t, entry) {
		t.Error("enabled without an autostart file")
	}
}

func TestEnablingWritesHiddenStartOfExecutable(t *testing.T) {
	configDir := t.TempDir()
	entry := autostart.NewEntry(configDir, "/home/user/.local/bin/meanwhile")

	mustSucceed(t, entry.SetEnabled(true))

	if !isEnabled(t, entry) {
		t.Fatal("not enabled after enabling")
	}
	if content := readEntry(t, configDir); !strings.Contains(content, "\nExec=\"/home/user/.local/bin/meanwhile\" --hidden\n") {
		t.Errorf("entry has no hidden start of the executable:\n%s", content)
	}
}

func TestDisablingRemovesFile(t *testing.T) {
	configDir := t.TempDir()
	entry := autostart.NewEntry(configDir, "/usr/bin/meanwhile")
	mustSucceed(t, entry.SetEnabled(true))

	mustSucceed(t, entry.SetEnabled(false))

	if isEnabled(t, entry) {
		t.Error("still enabled after disabling")
	}
	mustSucceed(t, entry.SetEnabled(false))
}

func TestEntryDisabledByDesktopIsReportedAndCanBeEnabledAgain(t *testing.T) {
	for _, disablingLine := range []string{"X-GNOME-Autostart-enabled=false", "Hidden=true"} {
		t.Run(disablingLine, func(t *testing.T) {
			configDir := t.TempDir()
			entry := autostart.NewEntry(configDir, "/usr/bin/meanwhile")
			mustSucceed(t, entry.SetEnabled(true))
			path := filepath.Join(configDir, "autostart", "io.github.yorik8877.meanwhile.desktop")
			mustSucceed(t, os.WriteFile(path, []byte(readEntry(t, configDir)+disablingLine+"\n"), 0o644))

			if isEnabled(t, entry) {
				t.Fatal("entry disabled by the desktop is reported as enabled")
			}
			mustSucceed(t, entry.SetEnabled(true))
			if !isEnabled(t, entry) {
				t.Error("entry stays disabled after enabling")
			}
		})
	}
}

func TestExecutablePathIsQuotedForDesktopEntry(t *testing.T) {
	configDir := t.TempDir()
	entry := autostart.NewEntry(configDir, `/home/my user/bin/stop"watch$100%`)

	mustSucceed(t, entry.SetEnabled(true))

	// The Exec quoting adds a backslash before " and $, the string escaping
	// of desktop entries then doubles every backslash; % becomes %%.
	want := `
Exec="/home/my user/bin/stop\\"watch\\$100%%" --hidden
`
	if content := readEntry(t, configDir); !strings.Contains(content, want) {
		t.Errorf("entry does not contain %q:\n%s", want, content)
	}
}

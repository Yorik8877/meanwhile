// Package autostart turns starting the program at login on and off through
// an XDG autostart desktop entry.
package autostart

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const entryFileName = "io.github.stopwatch.desktop"

const entryTemplate = `[Desktop Entry]
Type=Application
Name=Stopwatch
Name[ru]=Секундомер
Comment=Track time spent on parallel tasks
Comment[ru]=Учёт времени по параллельным задачам
Exec=%s --hidden
Icon=io.github.stopwatch
Terminal=false
X-GNOME-Autostart-enabled=true
`

// Entry is the autostart desktop entry of one executable.
type Entry struct {
	path       string
	executable string
}

// NewEntry places the entry in configDir/autostart, where configDir is
// $XDG_CONFIG_HOME or ~/.config.
func NewEntry(configDir, executable string) *Entry {
	return &Entry{path: filepath.Join(configDir, "autostart", entryFileName), executable: executable}
}

// Enabled reports false both for a missing entry and for an entry that the
// desktop's own startup settings switched off without deleting it.
func (entry *Entry) Enabled() (bool, error) {
	file, err := os.Open(entry.path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("open autostart entry: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if isDisablingLine(scanner.Text()) {
			return false, nil
		}
	}
	return true, scanner.Err()
}

func (entry *Entry) SetEnabled(enabled bool) error {
	if !enabled {
		if err := os.Remove(entry.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove autostart entry: %w", err)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(entry.path), 0o755); err != nil {
		return fmt.Errorf("create autostart directory: %w", err)
	}
	content := fmt.Sprintf(entryTemplate, quoteExecArgument(entry.executable))
	if err := os.WriteFile(entry.path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write autostart entry: %w", err)
	}
	return nil
}

func isDisablingLine(line string) bool {
	key, value, found := strings.Cut(line, "=")
	if !found {
		return false
	}
	key, value = strings.TrimSpace(key), strings.TrimSpace(value)
	return (key == "X-GNOME-Autostart-enabled" && value == "false") || (key == "Hidden" && value == "true")
}

// quoteExecArgument follows the desktop entry specification: inside double
// quotes ", `, $ and \ get a backslash, then the string escaping of values
// doubles every backslash, and % is written as %% so that it is not taken
// for a field code.
func quoteExecArgument(argument string) string {
	var quoted strings.Builder
	for _, character := range argument {
		if strings.ContainsRune("\"`$\\", character) {
			quoted.WriteRune('\\')
		}
		quoted.WriteRune(character)
	}
	escaped := strings.ReplaceAll(quoted.String(), `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, "%", "%%")
	return `"` + escaped + `"`
}

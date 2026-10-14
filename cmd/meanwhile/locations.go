package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const applicationDirectoryName = "meanwhile"

type locations struct {
	DataDir    string
	StateDir   string
	RuntimeDir string
	// ConfigHome holds the autostart entry. It is empty for a development
	// copy, which must not touch the installed program's autostart.
	ConfigHome string
}

// resolveLocations follows the XDG base directory specification. An override
// puts everything into one directory, so that a development build can run
// next to the installed program without sharing its data or its instance lock.
func resolveLocations(environment func(string) string, homeDir, override string) locations {
	if override != "" {
		return locations{DataDir: override, StateDir: override, RuntimeDir: override}
	}
	dataHome := absoluteOr(environment("XDG_DATA_HOME"), filepath.Join(homeDir, ".local", "share"))
	stateHome := absoluteOr(environment("XDG_STATE_HOME"), filepath.Join(homeDir, ".local", "state"))
	stateDir := filepath.Join(stateHome, applicationDirectoryName)
	return locations{
		DataDir:    filepath.Join(dataHome, applicationDirectoryName),
		StateDir:   stateDir,
		RuntimeDir: absoluteOr(environment("XDG_RUNTIME_DIR"), stateDir),
		ConfigHome: absoluteOr(environment("XDG_CONFIG_HOME"), filepath.Join(homeDir, ".config")),
	}
}

// absoluteOr ignores relative paths, as the XDG specification requires.
func absoluteOr(value, fallback string) string {
	if filepath.IsAbs(value) {
		return value
	}
	return fallback
}

func prepareLocations(override string) (locations, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return locations{}, fmt.Errorf("find home directory: %w", err)
	}
	resolved := resolveLocations(os.Getenv, homeDir, override)
	for _, directory := range []string{resolved.DataDir, resolved.StateDir, resolved.RuntimeDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return locations{}, fmt.Errorf("create %s: %w", directory, err)
		}
	}
	return resolved, nil
}

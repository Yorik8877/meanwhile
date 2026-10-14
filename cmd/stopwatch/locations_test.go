package main

import "testing"

func environmentOf(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func TestResolveLocationsUsesDefaultsUnderHome(t *testing.T) {
	resolved := resolveLocations(environmentOf(nil), "/home/user", "")

	want := locations{
		DataDir:    "/home/user/.local/share/stopwatch",
		StateDir:   "/home/user/.local/state/stopwatch",
		RuntimeDir: "/home/user/.local/state/stopwatch",
		ConfigHome: "/home/user/.config",
	}
	if resolved != want {
		t.Errorf("locations = %+v, want %+v", resolved, want)
	}
}

func TestResolveLocationsHonoursXDGVariables(t *testing.T) {
	environment := environmentOf(map[string]string{
		"XDG_DATA_HOME":   "/data",
		"XDG_STATE_HOME":  "/state",
		"XDG_RUNTIME_DIR": "/run/user/1000",
		"XDG_CONFIG_HOME": "/config",
	})

	resolved := resolveLocations(environment, "/home/user", "")

	want := locations{DataDir: "/data/stopwatch", StateDir: "/state/stopwatch", RuntimeDir: "/run/user/1000", ConfigHome: "/config"}
	if resolved != want {
		t.Errorf("locations = %+v, want %+v", resolved, want)
	}
}

func TestResolveLocationsIgnoresRelativeXDGPaths(t *testing.T) {
	environment := environmentOf(map[string]string{"XDG_DATA_HOME": "relative/data", "XDG_RUNTIME_DIR": "run"})

	resolved := resolveLocations(environment, "/home/user", "")

	if resolved.DataDir != "/home/user/.local/share/stopwatch" || resolved.RuntimeDir != "/home/user/.local/state/stopwatch" {
		t.Errorf("locations = %+v, want defaults", resolved)
	}
}

// A development copy has no ConfigHome, so it cannot replace the autostart
// entry of the installed program.
func TestDataDirOverrideHoldsEverything(t *testing.T) {
	resolved := resolveLocations(environmentOf(map[string]string{"XDG_RUNTIME_DIR": "/run/user/1000"}), "/home/user", "/tmp/dev")

	want := locations{DataDir: "/tmp/dev", StateDir: "/tmp/dev", RuntimeDir: "/tmp/dev"}
	if resolved != want {
		t.Errorf("locations = %+v, want %+v", resolved, want)
	}
}

package plugin

import (
	"encoding/json"
	"strings"
)

// ThemeStorageState represents the zustand-persisted theme state from CLIProxyAPI Management Center.
type ThemeStorageState struct {
	State struct {
		Theme         string `json:"theme"`
		ResolvedTheme string `json:"resolvedTheme"`
	} `json:"state"`
	Version int `json:"version"`
}

// ResolveThemeFromStorage extracts theme configuration from the 'cli-proxy-theme' localStorage JSON.
// Returns the user-selected theme setting ("auto", "white", "light", "dark") and the resolved theme ("dark" or "light"/"white").
func ResolveThemeFromStorage(storageJSON string) (theme string, resolved string) {
	theme = "auto"
	resolved = "light"

	s := strings.TrimSpace(storageJSON)
	if s == "" {
		return theme, resolved
	}

	// 1. Try zustand persist format
	var zustandState ThemeStorageState
	if err := json.Unmarshal([]byte(s), &zustandState); err == nil && (zustandState.State.Theme != "" || zustandState.State.ResolvedTheme != "") {
		if zustandState.State.Theme != "" {
			theme = zustandState.State.Theme
		}
		if zustandState.State.ResolvedTheme != "" {
			resolved = zustandState.State.ResolvedTheme
		} else {
			switch theme {
			case "dark":
				resolved = "dark"
			case "white", "light":
				resolved = "light"
			default:
				resolved = "light"
			}
		}
		return theme, resolved
	}

	// 2. Try simple map {"theme": "..."}
	var simpleMap map[string]any
	if err := json.Unmarshal([]byte(s), &simpleMap); err == nil {
		if t, ok := simpleMap["theme"].(string); ok && t != "" {
			theme = t
			if r, ok2 := simpleMap["resolvedTheme"].(string); ok2 && r != "" {
				resolved = r
			} else if theme == "dark" {
				resolved = "dark"
			} else {
				resolved = "light"
			}
			return theme, resolved
		}
	}

	// 3. Fallback for raw string literals like `"dark"` or `dark`
	clean := strings.ToLower(strings.Trim(s, "\" \t\r\n"))
	switch clean {
	case "dark":
		return "dark", "dark"
	case "white":
		return "white", "white"
	case "light":
		return "light", "light"
	case "auto":
		return "auto", "light"
	}

	return theme, resolved
}

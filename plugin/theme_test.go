package plugin

import (
	"testing"
)

func TestResolveThemeFromStorage(t *testing.T) {
	tests := []struct {
		name         string
		storageJSON  string
		wantTheme    string
		wantResolved string
	}{
		{
			name:         "zustand dark theme",
			storageJSON:  `{"state":{"theme":"dark","resolvedTheme":"dark"},"version":0}`,
			wantTheme:    "dark",
			wantResolved: "dark",
		},
		{
			name:         "zustand white theme",
			storageJSON:  `{"state":{"theme":"white","resolvedTheme":"light"},"version":0}`,
			wantTheme:    "white",
			wantResolved: "light",
		},
		{
			name:         "zustand light theme",
			storageJSON:  `{"state":{"theme":"light","resolvedTheme":"light"},"version":0}`,
			wantTheme:    "light",
			wantResolved: "light",
		},
		{
			name:         "zustand auto theme with resolved dark",
			storageJSON:  `{"state":{"theme":"auto","resolvedTheme":"dark"},"version":0}`,
			wantTheme:    "auto",
			wantResolved: "dark",
		},
		{
			name:         "zustand auto theme with resolved light",
			storageJSON:  `{"state":{"theme":"auto","resolvedTheme":"light"},"version":0}`,
			wantTheme:    "auto",
			wantResolved: "light",
		},
		{
			name:         "simple json dark",
			storageJSON:  `{"theme":"dark"}`,
			wantTheme:    "dark",
			wantResolved: "dark",
		},
		{
			name:         "simple json white",
			storageJSON:  `{"theme":"white"}`,
			wantTheme:    "white",
			wantResolved: "light",
		},
		{
			name:         "raw string dark",
			storageJSON:  `"dark"`,
			wantTheme:    "dark",
			wantResolved: "dark",
		},
		{
			name:         "raw string white",
			storageJSON:  `white`,
			wantTheme:    "white",
			wantResolved: "white",
		},
		{
			name:         "empty string defaults to auto/light",
			storageJSON:  "",
			wantTheme:    "auto",
			wantResolved: "light",
		},
		{
			name:         "invalid json defaults to auto/light",
			storageJSON:  `{not-valid-json`,
			wantTheme:    "auto",
			wantResolved: "light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTheme, gotResolved := ResolveThemeFromStorage(tt.storageJSON)
			if gotTheme != tt.wantTheme {
				t.Errorf("theme = %q, want %q", gotTheme, tt.wantTheme)
			}
			if gotResolved != tt.wantResolved {
				t.Errorf("resolved = %q, want %q", gotResolved, tt.wantResolved)
			}
		})
	}
}

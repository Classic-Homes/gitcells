package updater

import (
	"testing"
)

func TestVersionComparison(t *testing.T) {
	u := New("dev")

	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
	}{
		{"dev version should always update", "v1.0.0", "dev", true},
		{"newer version available", "v1.2.0", "v1.1.0", true},
		{"same version", "v1.1.0", "v1.1.0", false},
		{"older version", "v1.0.0", "v1.1.0", false},
		{"prerelease to stable", "v1.1.0", "v1.1.0-beta.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := u.isNewerVersion(tt.latest, tt.current); got != tt.expected {
				t.Errorf("isNewerVersion(%s, %s) = %v, expected %v",
					tt.latest, tt.current, got, tt.expected)
			}
		})
	}
}

func TestNewWithPrerelease(t *testing.T) {
	if New("1.0.0").AllowPrerelease {
		t.Error("default constructor should not allow prereleases")
	}
	if !NewWithPrerelease("1.0.0", true).AllowPrerelease {
		t.Error("NewWithPrerelease(true) should set AllowPrerelease")
	}
	if NewWithPrerelease("1.0.0", false).AllowPrerelease {
		t.Error("NewWithPrerelease(false) should not set AllowPrerelease")
	}
}

func TestReleasePageURL(t *testing.T) {
	if ReleasePageURL() == "" {
		t.Error("ReleasePageURL returned empty string")
	}
}

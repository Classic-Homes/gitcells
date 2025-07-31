package updater

import (
	"testing"
)

func TestVersionComparison(t *testing.T) {
	updater := New("dev")
	
	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
	}{
		{
			name:     "dev version should always update",
			latest:   "v1.0.0",
			current:  "dev",
			expected: true,
		},
		{
			name:     "newer version available",
			latest:   "v1.2.0",
			current:  "v1.1.0",
			expected: true,
		},
		{
			name:     "same version",
			latest:   "v1.1.0",
			current:  "v1.1.0",
			expected: false,
		},
		{
			name:     "older version",
			latest:   "v1.0.0",
			current:  "v1.1.0",
			expected: false,
		},
		{
			name:     "prerelease to stable",
			latest:   "v1.1.0",
			current:  "v1.1.0-beta.1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updater.isNewerVersion(tt.latest, tt.current)
			if result != tt.expected {
				t.Errorf("isNewerVersion(%s, %s) = %v, expected %v", 
					tt.latest, tt.current, result, tt.expected)
			}
		})
	}
}

func TestGetAssetName(t *testing.T) {
	updater := New("1.0.0")
	assetName := updater.getAssetName()
	
	// Should contain OS and architecture
	if assetName == "" {
		t.Error("getAssetName() returned empty string")
	}
	
	// Should be in format os_arch
	if len(assetName) < 5 { // minimum like "linux_amd64" would be longer
		t.Errorf("getAssetName() returned unexpectedly short name: %s", assetName)
	}
}

func TestVerifyChecksum(t *testing.T) {
	updater := New("1.0.0")
	
	testData := []byte("test data")
	// SHA256 of "test data"
	expectedChecksum := "916f0027a575074ce72a331777c3478d6513f786a591bd892da1a577bf2335f9"
	
	err := updater.VerifyChecksum(testData, expectedChecksum)
	if err != nil {
		t.Errorf("VerifyChecksum() failed with correct checksum: %v", err)
	}
	
	// Test with wrong checksum
	wrongChecksum := "wrong_checksum"
	err = updater.VerifyChecksum(testData, wrongChecksum)
	if err == nil {
		t.Error("VerifyChecksum() should have failed with wrong checksum")
	}
}

func TestNewWithPrerelease(t *testing.T) {
	tests := []struct {
		name           string
		allowPrerelease bool
		expected       bool
	}{
		{
			name:           "default constructor should not allow prerelease",
			allowPrerelease: false,
			expected:       false,
		},
		{
			name:           "explicit constructor should allow prerelease when set",
			allowPrerelease: true,
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updater *Updater
			if tt.name == "default constructor should not allow prerelease" {
				updater = New("1.0.0")
			} else {
				updater = NewWithPrerelease("1.0.0", tt.allowPrerelease)
			}
			
			if updater.AllowPrerelease != tt.expected {
				t.Errorf("AllowPrerelease = %v, expected %v", updater.AllowPrerelease, tt.expected)
			}
		})
	}
}
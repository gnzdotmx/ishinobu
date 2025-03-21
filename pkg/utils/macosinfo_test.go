package utils

import (
	"strings"
	"testing"
)

func TestGetMacOSVersion(t *testing.T) {
	// Test successful case
	version, err := GetMacOSVersion()
	if err != nil {
		t.Errorf("GetMacOSVersion() failed: %v", err)
	}

	// Check if version string is not empty
	if version == "" {
		t.Error("GetMacOSVersion() returned empty string")
	}

	// Check if version follows expected format (e.g., "13.1" or "14.2.1")
	parts := strings.Split(version, ".")
	if len(parts) < 2 || len(parts) > 3 {
		t.Errorf("GetMacOSVersion() returned unexpected format: %s", version)
	}
}

func TestGetHostname(t *testing.T) {
	// Test successful case
	hostname, err := GetHostname()
	if err != nil {
		t.Errorf("GetHostname() failed: %v", err)
	}

	// Check if hostname string is not empty
	if hostname == "" {
		t.Error("GetHostname() returned empty string")
	}

	// Check if hostname contains any invalid characters
	invalidChars := []string{" ", "\n", "\t", "\r"}
	for _, char := range invalidChars {
		if strings.Contains(hostname, char) {
			t.Errorf("GetHostname() returned hostname with invalid character: %s", hostname)
		}
	}
}

package utils

import (
	"os/exec"
	"strings"
)

func GetMacOSVersion() (string, error) {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func GetHostname() (string, error) {
	out, err := exec.Command("hostname").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

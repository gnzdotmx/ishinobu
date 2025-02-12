package utils

import (
	"os/exec"
	"strings"
)

func ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

package utils

import (
	"os"
	"os/exec"
	"strings"
)

// ExecCommand is a variable that can be replaced in tests to mock command execution
var ExecCommand = func(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	return cmd.Output()
}

func ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// ExecuteCommandWithEnv is a variable that can be replaced in tests to mock command execution with env
var ExecuteCommandWithEnv = func(command string, env []string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

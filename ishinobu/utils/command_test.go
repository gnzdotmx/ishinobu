package utils

import (
	"strings"
	"testing"
)

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "echo command",
			command:    "echo",
			args:       []string{"hello"},
			wantOutput: "hello",
			wantErr:    false,
		},
		{
			name:       "multiple args",
			command:    "echo",
			args:       []string{"hello", "world"},
			wantOutput: "hello world",
			wantErr:    false,
		},
		{
			name:       "invalid command",
			command:    "nonexistentcommand",
			args:       []string{},
			wantOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExecuteCommand(tt.command, tt.args...)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && strings.TrimSpace(got) != tt.wantOutput {
				t.Errorf("ExecuteCommand() = %v, want %v", got, tt.wantOutput)
			}
		})
	}
}

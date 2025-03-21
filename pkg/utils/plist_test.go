package utils

import (
	"testing"
)

func TestParseBiPList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid plist",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>TestKey</key>
	<string>TestValue</string>
	<key>NumberKey</key>
	<integer>42</integer>
</dict>
</plist>`,
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid plist",
			input:   "not a valid plist",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBiPList(tt.input)

			// Check error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBiPList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For valid cases, verify the parsed content
			if !tt.wantErr {
				if result == nil {
					t.Error("ParseBiPList() returned nil result for valid input")
					return
				}

				// Check if expected keys exist and have correct values
				if val, ok := result["TestKey"].(string); !ok || val != "TestValue" {
					t.Errorf("ParseBiPList() TestKey = %v, want %v", val, "TestValue")
				}

				if val, ok := result["NumberKey"].(uint64); !ok || val != 42 {
					t.Errorf("ParseBiPList() NumberKey = %v, want %v", val, 42)
				}
			}
		})
	}
}

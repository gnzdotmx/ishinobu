package utils

import (
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	result := Now()
	_, err := time.Parse(TimeFormat, result)
	if err != nil {
		t.Errorf("Now() returned an invalid timestamp format: %v", err)
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid timestamp",
			input:   "2023-01-02 15:04:05.123456-0700",
			want:    "2023-01-02T15:04:05-07:00",
			wantErr: false,
		},
		{
			name:    "invalid timestamp",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimestamp(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertDateString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid date string",
			input:   "Oct 26 19:34:13",
			wantErr: false,
		},
		{
			name:    "invalid date string",
			input:   "Invalid date",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertDateString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertDateString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				_, err := time.Parse(TimeFormat, got)
				if err != nil {
					t.Errorf("ConvertDateString() returned invalid timestamp format: %v", err)
				}
			}
		})
	}
}

func TestConvertCFAbsoluteTimeToDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid CF time",
			input:   "695569202",
			want:    "2023-01-16T13:40:02Z",
			wantErr: false,
		},
		{
			name:    "invalid CF time",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertCFAbsoluteTimeToDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertCFAbsoluteTimeToDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertCFAbsoluteTimeToDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseChromeTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid chrome timestamp",
			input: "13303766907771771",
			want:  "2023-01-02T15:04:05Z",
		},
		{
			name:  "zero timestamp",
			input: "0",
			want:  "",
		},
		{
			name:  "invalid timestamp",
			input: "invalid",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseChromeTimestamp(tt.input)
			if tt.want == "" {
				if got != "" {
					t.Errorf("ParseChromeTimestamp() = %v, want empty string", got)
				}
			} else {
				_, err := time.Parse(TimeFormat, got)
				if err != nil {
					t.Errorf("ParseChromeTimestamp() returned invalid timestamp format: %v", err)
				}
			}
		})
	}
}

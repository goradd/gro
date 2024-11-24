package query

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "Short string",
			input:    "123",
			expected: time.Time{}, // Should return zero time
		},
		{
			name:     "Unix time",
			input:    "1637689932",
			expected: time.Unix(1637689932, 0).UTC(),
		},
		{
			name:     "Unix time with fractional seconds",
			input:    "1637689932.123456789",
			expected: time.Unix(1637689932, 123456789).UTC(),
		},
		{
			name:     "RFC3339 format",
			input:    "2023-11-01T12:34:56Z",
			expected: time.Date(2023, 11, 1, 12, 34, 56, 0, time.UTC),
		},
		{
			name:     "Date only",
			input:    "2023-11-01",
			expected: time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Date and time",
			input:    "2023-11-01 12:34:56",
			expected: time.Date(2023, 11, 1, 12, 34, 56, 0, time.UTC),
		},
		{
			name:     "Date, time, and timezone",
			input:    "2023-11-01 12:34:56 +0700",
			expected: time.Date(2023, 11, 1, 12, 34, 56, 0, time.FixedZone("", 7*3600)).UTC(),
		},
		{
			name:     "Date, time, timezone, and locale",
			input:    "2023-11-01 12:34:56 +0700 MST",
			expected: time.Date(2023, 11, 1, 12, 34, 56, 0, time.FixedZone("MST", 7*3600)).UTC(),
		},
		{
			name:     "Time only",
			input:    "12:34:56",
			expected: time.Date(0, 1, 1, 12, 34, 56, 0, time.UTC),
		},
		{
			name:     "Invalid format",
			input:    "not a time",
			expected: time.Time{}, // Should return zero time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTime(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("ParseTime(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

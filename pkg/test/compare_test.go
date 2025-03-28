package test

import (
	"testing"
)

func TestEqualDecimals(t *testing.T) {
	tests := []struct {
		d1, d2   string
		expected bool
	}{
		{"001.2300", "1.23", true},
		{"+000.000", "0", true},
		{"123", "0123.0", true},
		{"1.20", "1.2", true},
		{"1.200", "1.2", true},
		{"1.23", "1.23001", false},
		{"-001.2300", "-1.23", true},
		{"-000.000", "0", true},   // treating -0 as 0
		{"0", "-0", true},         // treating -0 as 0
		{"+000.00", "-0.0", true}, // treating -0 as 0
		{"-123", "-0123.0", true},
		{"1.23", "-1.23", false},
		{"", "0", true},        // empty string treated as zero
		{"+.3", "0.30", true},  // empty string treated as zero
		{"-.3", "-0.30", true}, // empty string treated as zero
	}

	for _, tt := range tests {
		t.Run(tt.d1+"=="+tt.d2, func(t *testing.T) {
			result := EqualDecimals(tt.d1, tt.d2)
			if result != tt.expected {
				t.Errorf("EqualDecimals(%q, %q) = %v; want %v", tt.d1, tt.d2, result, tt.expected)
			}
		})
	}
}

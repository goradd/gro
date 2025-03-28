package test

import "strings"

// EqualDecimals returns true if d1 and d2 are equal decimal strings.
// Allows for the strings to have a leading + and any number of leading zeros, or trailing zeros after a decimal point.
func EqualDecimals(d1, d2 string) bool {
	return normalizeDecimal(d1) == normalizeDecimal(d2)
}

func normalizeDecimal(s string) string {
	if s == "" {
		return "0" // treat empty string as zero
	}

	sign := ""
	// Handle sign
	if s[0] == '+' {
		s = s[1:]
	} else if s[0] == '-' {
		sign = "-"
		s = s[1:]
	}

	// Split into integer and fractional parts
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	// Trim leading zeros from integer part
	intPart = strings.TrimLeft(intPart, "0")
	if intPart == "" {
		intPart = "0"
	}

	// Trim trailing zeros from fractional part
	fracPart = strings.TrimRight(fracPart, "0")

	// If both parts are effectively zero, return "0"
	if intPart == "0" && fracPart == "" {
		return "0"
	}

	// Reconstruct normalized number
	if fracPart != "" {
		return sign + intPart + "." + fracPart
	}
	return sign + intPart
}

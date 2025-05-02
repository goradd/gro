package schema

import (
	"strings"
	"unicode"
)

// reservedWords is a set of Go keywords.
var reservedWords = map[string]struct{}{
	"break": {}, "default": {}, "func": {}, "interface": {}, "select": {},
	"case": {}, "defer": {}, "go": {}, "map": {}, "struct": {},
	"chan": {}, "else": {}, "goto": {}, "package": {}, "switch": {},
	"const": {}, "fallthrough": {}, "if": {}, "range": {}, "type": {},
	"continue": {}, "for": {}, "import": {}, "return": {}, "var": {},
}

// SanitizePackageName returns a string that is valid as a Go package name.
// It only allows lowercase letters and digits, replacing all other characters with underscores.
// Uppercase letters are converted to lowercase.
// If the result starts with a non-letter or is a reserved keyword, it prepends an underscore.
func SanitizePackageName(input string) string {
	var b strings.Builder

	for _, r := range input {
		switch {
		case unicode.IsLower(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case unicode.IsUpper(r):
			b.WriteRune(unicode.ToLower(r))
		default:
			b.WriteRune('_')
		}
	}

	result := b.String()

	if result == "" || !unicode.IsLower(rune(result[0])) {
		result = "_" + result
	}

	if _, isReserved := reservedWords[result]; isReserved {
		result = "_" + result
	}

	return result
}

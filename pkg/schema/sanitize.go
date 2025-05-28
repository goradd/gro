package schema

import (
	"log/slog"
	"strings"
	"unicode"
)

// reservedWords is a set of Go keywords and other reserved words.
var reservedWords = map[string]struct{}{
	// go keywords
	"break": {}, "default": {}, "func": {}, "interface": {}, "select": {},
	"case": {}, "defer": {}, "go": {}, "map": {}, "struct": {},
	"chan": {}, "else": {}, "goto": {}, "package": {}, "switch": {},
	"const": {}, "fallthrough": {}, "if": {}, "range": {}, "type": {},
	"continue": {}, "for": {}, "import": {}, "return": {}, "var": {},
	// go built-in types
	"int": {}, "string": {}, "bool": {}, "byte": {}, "rune": {}, "error": {},
	// goradd function names that will get code generated
	"key": {}, "label": {}, "copy": {}, "primary_key": {},
	// There are others that are unlikely or technically allowed but might conflict. Add as needed.
}

// SanitizeName returns a string that is valid as a Go identifier.
// It only allows lowercase letters and digits, replacing all other characters with underscores.
// Uppercase letters are converted to lowercase.
// If the result starts with a non-letter or is a reserved keyword, it prepends an underscore.
func SanitizeName(input string) string {
	var b strings.Builder

	for _, r := range input {
		switch {
		case unicode.IsLower(r) || unicode.IsDigit(r) || r == '_':
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

	if input != result {
		slog.Warn("Name was modified because its not a valid go identifier",
			slog.String("old name", input),
			slog.String("new name", result))
	}

	return result
}

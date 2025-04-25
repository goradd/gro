package schema

import (
	"encoding/json"
	"github.com/goradd/goradd/pkg/stringmap"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"slices"
	"strings"
)

const ConstKey = "const"
const LabelKey = "label"
const IdentifierKey = "identifier"

type EnumField struct {
	// Identifier is the name used in Go code to access the data.
	Identifier string `json:"identifier,omitempty"`
	// IdentifierPlural is the plural of the Identifier.
	IdentifierPlural string `json:"identifier_plural,omitempty"`
	// Type is the type of the column.
	// It is inferred from the data, but can be specifically identified if needed, like if integer data
	// should be presented as floating point.
	Type ColumnType `json:"type"`
}

// EnumTable describes a table that contains enumerated values. The resulting Go code will be a type
// with constants for each value in the database. An EnumTable is processed with its data at compile time
// and cannot be modified by the application. The values are stored in the database, but are not accessed
// by database queries.
type EnumTable struct {
	// Name is the name of the table in the database.
	// The name should have the Database.EnumTableSuffix value as a suffix.
	Name string `json:"name"`
	// For databases that support schemas, this is the name of the schema of the table.
	// Leave blank for the default schema.
	// Databases that do not support schemas will have this prepended to the name of the table.
	Schema string `json:"schema,omitempty"`

	// Values are the enum values themselves.
	// Each entry must have the same key-value pairs.
	// Each entry must have a "const" key with an integer value. That will become the const value in Go code.
	// Each entry must have either a "label" key, an "identifier" key, or both. The label is a human-readable
	// description of the value, and identifier is a snake_case equivalent used as the JSON identifier for the value.
	// If either is missing, it will be generated from the other.
	// Additional entries will create accessor functions for those values.
	// The first entry will determine what types are inferred for each value.
	Values []map[string]any `json:"values"`

	// Fields provides further descriptions for the accessor functions.
	// If the entries are omitted, a default will be generated.
	// Keys are the same as the keys in the Values entries.
	Fields map[string]EnumField `json:"fields,omitempty"`

	// Label is the name of the object when describing it to humans.
	// If creating a multi-language app, your app would provide translation from this string to the language of choice.
	// Can be multiple words. Should be lower-case. The app will use github.com/goradd/strings.ReverseLabel() to capitalize this if needed.
	// If left blank, the app will base this on the Name of the table.
	Label string `json:"label,omitempty"`

	// LabelPlural is the plural form of the Label.
	LabelPlural string `json:"label_plural,omitempty"`

	// Identifier is the corresponding Go object name. It must obey Go identifier labeling rules. Leave blank
	// to base it on the Name.
	Identifier string `json:"identifier,omitempty"`

	// IdentifierPlural is the plural form of Identifier.
	IdentifierPlural string `json:"identifier_plural,omitempty"`

	// Key is used internally to aid in database synchronization.
	Key string `json:"key,omitempty"`
}

// QualifiedName returns the name of the table in the database, including its schema if applicable.
func (t *EnumTable) QualifiedName() string {
	if t.Schema == "" {
		return t.Name
	} else {
		return t.Schema + "." + t.Name
	}
}

// FillDefaults will fill in empty values in the EnumTable struct based on the values provided.
func (t *EnumTable) FillDefaults(suffix string) {
	name := strings.TrimSuffix(t.QualifiedName(), suffix)
	if t.Label == "" {
		t.Label = strings2.Title(name)
	}
	if t.LabelPlural == "" {
		t.LabelPlural = strings2.Plural(t.Label)
	}
	if t.Identifier == "" {
		t.Identifier = snaker.SnakeToCamelIdentifier(name)
	}
	if t.IdentifierPlural == "" {
		t.IdentifierPlural = strings2.Plural(t.Identifier)
	}
	if len(t.Values) == 0 {
		return // this will not generate anything. Assume it is a placeholder to be filled in later by the developer.
	}
	// Sanity check the values
	for _, v := range t.Values {
		if _, ok := v[ConstKey]; !ok {
			panic("const is required for every value entry")
		}
		_, hasId := v[IdentifierKey].(string)
		_, hasLabel := v[LabelKey].(string)
		if !(hasId || hasLabel) {
			panic("A label or identifier of type string is required for every value entry")
		}
		if !hasId {
			v[IdentifierKey] = snaker.CamelToSnakeIdentifier(v[LabelKey].(string))
		}
		if !hasLabel {
			v[LabelKey] = strings2.Title(v[IdentifierKey].(string))
		}
	}

	for _, k := range t.FieldKeys() {
		f, ok := t.Fields[k]
		if !ok {
			f = EnumField{}
		}

		if f.Identifier == "" {
			f.Identifier = snaker.SnakeToCamelIdentifier(k)
		}
		if f.IdentifierPlural == "" {
			f.IdentifierPlural = strings2.Plural(f.Identifier)
		}
		if f.Type == ColTypeUnknown {
			switch v := t.Values[0][k].(type) {
			case string:
				f.Type = ColTypeString
			case int:
			case int64:
			case int32:
				f.Type = ColTypeInt
			case float64:
			case float32:
				f.Type = ColTypeFloat
			case json.Number:
				if _, err := v.Int64(); err == nil {
					f.Type = ColTypeInt
				} else {
					f.Type = ColTypeFloat
				}
			}
		}
		t.Fields[k] = f
	}
}

// FieldKeys returns the keys of the fields in deterministic order, with const, label and identifier first
func (t *EnumTable) FieldKeys() (keys []string) {
	keys = []string{"const", "label", "identifier"}
	if len(t.Values) == 0 {
		return
	}
	for _, k := range stringmap.SortedKeys(t.Values[0]) {
		if !slices.Contains(keys, k) {
			keys = append(keys, k)
		}
	}
	return
}

package schema

import (
	"encoding/json"
	"fmt"
	"github.com/goradd/goradd/pkg/stringmap"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"log/slog"
	"strings"
	"time"
)

const ValueKey = "value"
const LabelKey = "label"
const NameKey = "name"
const KeyKey = "key"

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

// EnumTable describes a table that contains enumerated values.
// The resulting Go code will be a named integer type with constants for each value.
// An EnumTable is processed with its data at compile time and cannot be modified by the application.
// The values are stored in the database, but are not accessed by database queries.
type EnumTable struct {
	// Name is the name of the table in the database.
	// By convention, the name should have the Database.EnumTableSuffix value as a suffix.
	// This suffix will be stripped off for items below that use Name as the basis for a default value.
	Name string `json:"name"`
	// For databases that support schemas, this is the name of the schema of the table.
	// Leave blank for the default schema.
	// Databases that do not support schemas will have this prepended to the name of the table.
	Schema string `json:"schema,omitempty"`

	// Values are the enum values themselves.
	// At a minimum, each entry must have a "name" field. This should be a CamelCase string that will be
	// used to build the constant value in the Go code.
	// By default, numbers for values will be assigned in order. If you set a "value" value, it will override this default.
	// By default, a "label" value will be assigned based on "name". This should be a Title Case string that will
	// represent the value to humans. Set "label" to override the default.
	// Alternatively specify a "key" value to be used as the json key and other appropriate contexts. This
	// should be a lower_snake_case value and will be generated if missing.
	// Additional entries will create accessor functions for those values. The type and identifiers will be
	// inferred from the entries, or you can use Fields to specify these values for the additional fields.
	Values []map[string]any `json:"values"`

	// Fields provides descriptions for additional values found in Values.
	// If the entries are omitted, a default will be generated.
	// Keys are the same as the keys in the Values entries.
	// "label" field will be generated automatically and should not be included on import.
	Fields map[string]EnumField `json:"fields,omitempty"`

	// Label is the name of the object when describing it to humans.
	// If creating a multi-language app, your app would provide translation from this string to the language of choice.
	// Can be multiple words. Should be Title Case.
	// If left blank, the app will base this on the Name of the table.
	Label string `json:"label,omitempty"`

	// LabelPlural is the plural form of the Label.
	LabelPlural string `json:"label_plural,omitempty"`

	// Identifier is the corresponding Go object name. It must obey Go identifier labeling rules.
	// Should be CamelCase.
	// Leave blank to base it on the Name.
	Identifier string `json:"identifier,omitempty"`

	// IdentifierPlural is the plural form of Identifier.
	IdentifierPlural string `json:"identifier_plural,omitempty"`

	// Comment is a place to put a comment in the JSON description file.
	// If the database driver supports it, it may be put in the database.
	Comment string `json:"comment,omitempty"`
}

// QualifiedTableName returns the name of the table in the database, including its schema if applicable.
func (t *EnumTable) QualifiedTableName() string {
	if t.Schema == "" {
		return t.Name
	} else {
		return t.Schema + "." + t.Name
	}
}

func (t *EnumTable) infer(db *Database) error {
	if t.Name == "" {
		return fmt.Errorf("enum table must have a name")
	}
	extraFields := make(map[string]EnumField)
	var i int
	for _, vMap := range t.Values {
		var hasValue, hasName bool
		for k, v := range vMap {
			switch k {
			case ValueKey:
				if v2, okN := v.(json.Number); okN {
					i2, err := v2.Int64()
					if err != nil {
						return fmt.Errorf(`"value" must be an integer: %w`, err)
					}
					vMap[ValueKey] = int(i2)
				} else if _, okI := v.(int); !okI {
					return fmt.Errorf(`"value" must be an integer: %v`, v)
				}
				hasValue = true
			case NameKey:
				if _, ok := v.(string); !ok {
					return fmt.Errorf(`"name" value is not a string: %v`, v)
				}
				hasName = true
			case KeyKey:
				if _, ok := v.(string); !ok {
					return fmt.Errorf(`"key" value is not a string: %v`, v)
				}
			case LabelKey:
				if _, ok := v.(string); !ok {
					return fmt.Errorf(`"label" value is not a string: %v`, v)
				}
				fallthrough
			default:
				identifier := snaker.SnakeToCamelIdentifier(k)
				t, v2 := inferColumnType(v)
				extraFields[k] = EnumField{
					Identifier:       identifier,
					IdentifierPlural: strings2.Plural(identifier),
					Type:             t,
				}
				vMap[k] = v2
			}
		}
		if !hasValue {
			i++
			vMap[ValueKey] = i
		}
		if !hasName {
			return fmt.Errorf(`"name" is a required field for an enum value and must be a string`)
		}
		for k, v := range extraFields {
			if _, ok := t.Fields[k]; !ok {
				t.Fields[k] = v
			}
		}
	}
	return nil
}

func inferColumnType(v any) (ColumnType, any) {
	if n, ok := v.(json.Number); ok {
		if i, err := n.Int64(); err != nil {
			return ColTypeInt, int(i)
		}
		if f, err := n.Float64(); err != nil {
			return ColTypeFloat, f
		}
		return ColTypeString, n.String()
	}
	if _, ok := v.(int); ok {
		return ColTypeInt, v
	}
	if s, ok := v.(string); ok {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			return ColTypeTime, t
		}
		return ColTypeString, s
	}
	if _, ok := v.(bool); ok {
		return ColTypeBool, v
	}
	return ColTypeUnknown, v
}

// fillDefaults will fill in empty values in the EnumTable struct based on the values provided.
func (t *EnumTable) fillDefaults(suffix string) {
	name := strings.TrimSuffix(t.QualifiedTableName(), suffix)
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
	if t.Fields == nil {
		t.Fields = make(map[string]EnumField)
	}
	for k, v := range t.Fields {
		if v.Identifier == "" {
			v.Identifier = snaker.SnakeToCamelIdentifier(k)
		}
		if v.IdentifierPlural == "" {
			v.IdentifierPlural = strings2.Plural(v.Identifier)
		}
		if v.Identifier == v.IdentifierPlural {
			slog.Warn("Enum field identifier is plural and should be singular.",
				slog.String("identifier", v.Identifier))
		}

		t.Fields[k] = v
	}
	t.Fields[LabelKey] = EnumField{
		Identifier:       "Label",
		IdentifierPlural: "Labels",
		Type:             ColTypeString,
	}
	for _, vMap := range t.Values {
		if _, ok := vMap[LabelKey]; !ok {
			vMap[LabelKey] = strings2.Title(vMap[NameKey].(string))
		}
		if _, ok := vMap[KeyKey]; !ok {
			vMap[KeyKey] = strings2.CamelToSnake(vMap[NameKey].(string))
		}
	}
}

// FieldKeys returns the keys of the fields in deterministic order, with label first
func (t *EnumTable) FieldKeys() (keys []string) {
	for _, k := range stringmap.SortedKeys(t.Fields) {
		if k == "label" {
			keys = append([]string{LabelKey}, keys...)
		} else {
			keys = append(keys, k)
		}
	}
	return
}

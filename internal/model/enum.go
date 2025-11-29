package model

import (
	"cmp"
	"log/slog"
	"slices"

	. "github.com/goradd/anyutil"
	"github.com/goradd/gro/internal/schema"
	. "github.com/goradd/gro/query"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
)

type ConstValue struct {
	// Value is the integer that will represent the value
	Value int
	// Name is the CamelCase name that represents the value (without the table name prefix)
	Name string
	// Key is the snake_case key that will represent the value.
	Key string
	// FieldValues are the additional values associated with the field. At a minimum, it will have a "label" entry.
	// This should correspond to the entries in Enum.Fields.
	FieldValues map[string]any
}

// Enum describes a structure that represents a fixed, enumerated type that will
// not change during the application's use. This will generate types that can be
// used as values for fields in the database.
type Enum struct {
	// DbKey is the key used to find the database in the global database cluster.
	DbKey string
	// QueryName is the name of the table to use in querying the database.
	QueryName string
	// Label is the english name of the object when describing it to the world.
	Label string
	// LabelPlural is the plural english name of the object.
	LabelPlural string
	// Identifier is the name of the item as a go type name.
	Identifier string
	// IdentifierPlural is the plural of the go type name.
	IdentifierPlural string
	// DecapIdentifier is the Identifier with the first letter lower case.
	DecapIdentifier string
	// Fields are the names of accessor functions associated with the type.
	// The first field name MUST be "label". There may be additional fields.
	Fields []EnumField
	// Constants are the constant identifiers that will be used for each enumerated value.
	// These are in ascending order by keys.
	Constants []ConstValue
}

func (tt *Enum) FieldQueryName(i int) string {
	return tt.Fields[i].QueryName
}

// FieldIdentifier returns the go name corresponding to the given field offset, or an empty string if out of bounds.
func (tt *Enum) FieldIdentifier(i int) string {

	return If(tt.Fields[i], tt.Fields[i].Identifier, "")
}

// FieldIdentifierPlural returns the go plural name corresponding to the given field offset, or an empty string if out of bounds.
func (tt *Enum) FieldIdentifierPlural(i int) string {
	return If(tt.Fields[i], tt.Fields[i].IdentifierPlural, "")
}

// FieldReceiverType returns the ReceiverType corresponding to the given field offset
func (tt *Enum) FieldReceiverType(i int) ReceiverType {
	return tt.Fields[i].Type
}

// FileName returns the default file name corresponding to the enum table.
func (tt *Enum) FileName() string {
	return snaker.CamelToSnake(tt.Identifier)
}

// newEnumTable will import the enum table from tableSchema.
// If an error occurs, the table will be returned with no Values.
func newEnumTable(dbKey string, enumSchema *schema.EnumTable) *Enum {
	t := &Enum{
		DbKey:            dbKey,
		QueryName:        enumSchema.QualifiedTableName(),
		Label:            enumSchema.Label,
		LabelPlural:      enumSchema.LabelPlural,
		Identifier:       enumSchema.Identifier,
		IdentifierPlural: enumSchema.IdentifierPlural,
		DecapIdentifier:  strings2.Decap(enumSchema.Identifier),
	}
	if len(enumSchema.Values) == 0 {
		slog.Error("Enum table " + t.QueryName + " has no Values entries. Specify constants by adding entries to this table schema.")
		return t
	}

	keys := enumSchema.FieldKeys()

	for _, k := range keys {
		f := EnumField{
			QueryName:        k,
			Identifier:       enumSchema.Fields[k].Identifier,
			IdentifierPlural: enumSchema.Fields[k].IdentifierPlural,
			Type:             ReceiverTypeFromSchema(enumSchema.Fields[k].Type, 0),
		}
		t.Fields = append(t.Fields, f)
	}

	for _, valueMap := range enumSchema.Values {
		c := ConstValue{
			Value:       valueMap[schema.ValueKey].(int),
			Name:        valueMap[schema.NameKey].(string),
			Key:         valueMap[schema.KeyKey].(string),
			FieldValues: make(map[string]any),
		}
		for k, v := range valueMap {
			switch k {
			case schema.ValueKey, schema.NameKey, schema.KeyKey: // already extracted above
			default:
				c.FieldValues[k] = v // includes label
			}
		}
		t.Constants = append(t.Constants, c)
	}
	slices.SortFunc(t.Constants, func(a, b ConstValue) int {
		return cmp.Compare(a.Value, b.Value)
	})
	return t
}

type EnumField struct {
	// QueryName is the name of the field in the database.
	// QueryNames should be lower_snake_case.
	QueryName string
	// Identifier is the name used in Go code to access the data.
	Identifier string
	// IdentifierPlural is the plural form of Identifier.
	IdentifierPlural string
	// Type is the ReceiverType of the column.
	Type ReceiverType
}

func (f EnumField) GoType() string {
	return f.Type.GoType()
}

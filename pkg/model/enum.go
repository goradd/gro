package model

import (
	"cmp"
	"fmt"
	. "github.com/goradd/anyutil"
	. "github.com/goradd/orm/pkg/query"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"log/slog"
	"slices"
)

type ConstVal struct {
	Value int
	Const string
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
	// Fields are the names of the fields defined in the table. The first field name MUST be the const field,
	// the 2nd the Label, and the third the Identifier. There may be additional fields.
	Fields []EnumField
	// Values are the go values that will be hardcoded and returned in accessor functions.
	// The map is keyed by row id, and then by field query name
	Values map[int]map[string]any
	// Constants are the constant identifiers that will be used for each enumerated value.
	// These are in ascending order by keys.
	Constants []ConstVal
}

// PkQueryName returns the name of the primary key field as used in database queries.
func (tt *Enum) PkQueryName() string {
	return tt.FieldQueryName(0)
}

func (tt *Enum) FieldQueryName(i int) string {
	return tt.Fields[i].QueryName
}

func (tt *Enum) FieldValue(row int, fieldNum int) any {
	name := tt.FieldQueryName(fieldNum)
	v := tt.Values[row][name]
	if IsNil(v) {
		v = tt.Fields[fieldNum].Type.DefaultValue()
	}
	return v
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
		QueryName:        enumSchema.QualifiedName(),
		Label:            enumSchema.Label,
		LabelPlural:      enumSchema.LabelPlural,
		Identifier:       enumSchema.Identifier,
		IdentifierPlural: enumSchema.IdentifierPlural,
		DecapIdentifier:  strings2.Decap(enumSchema.Identifier),
		Values:           make(map[int]map[string]any),
	}
	if len(enumSchema.Values) == 0 {
		slog.Error("Enum table " + t.QueryName + " has no Values entries. Specify constants by adding entries to this table schema.")
		return t
	}
	if len(enumSchema.Fields) < 2 {
		slog.Error("Enum table " + t.QueryName + " does not have at least 2 Fields entries. Specify fields by adding Fields to this table schema.")
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

	for i, row := range enumSchema.Values {
		if len(row) != len(t.Fields) {
			slog.Error(fmt.Sprintf("Enum table %s, Values row %d, does not have the same number of values as Fields.", t.QueryName, i))
			clear(t.Values)
			return t
		}

		valueMap := make(map[string]any)
		for _, k := range keys {
			valueMap[k] = row[k]
		}
		t.Constants = append(t.Constants, ConstVal{row[schema.ConstKey].(int), enumValueToConstant(t.Identifier, row[schema.IdentifierKey].(string))})
		t.Values[row[schema.ConstKey].(int)] = valueMap
	}
	slices.SortFunc(t.Constants, func(a, b ConstVal) int {
		return cmp.Compare(a.Value, b.Value)
	})
	return t
}

func enumValueToConstant(prefix string, v string) string {
	v = snaker.ForceCamelIdentifier(v)
	return prefix + v
}

type EnumField struct {
	// QueryName is the name of the field in the database.
	// The name of the first field is "const" by convention.
	// The name of the second field is "label".
	// The name of the third field is "identifier".
	// Additional fields are optional.
	// QueryNames should be lower_snake_case.
	QueryName string
	// Identifier is the name used in Go code to access the data.
	Identifier string
	// IdentifierPlural is the plural form of Identifier.
	IdentifierPlural string
	// Type is the type of the column.
	// The const column must be type ColTypeInt.
	// The label and identifier columns must be type ColTypeString.
	Type ReceiverType
}

func (f EnumField) GoType() string {
	return f.Type.GoType()
}

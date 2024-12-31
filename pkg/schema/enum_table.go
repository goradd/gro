package schema

import (
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"strings"
)

type EnumField struct {
	// Name is the name of the field in the database.
	// The name of the first field is typically "id" by convention.
	// The name of the second field must be "name".
	// The name of the following fields is up to you, but should be lower_snake_case.
	Name string `json:"name"`
	// Title is the title of the data stored in the field.
	Title string `json:"title,omitempty"`
	// TitlePlural is the plural of the Title.
	TitlePlural string `json:"title_plural,omitempty"`
	// Identifier is the name used in Go code to access the data.
	Identifier string `json:"identifier,omitempty"`
	// IdentifierPlural is the plural of the Identifier.
	IdentifierPlural string `json:"identifier_plural,omitempty"`
	// Type is the type of the column.
	// The first column must be type ColTypeInt.
	// The second column must be type ColTypeString.
	// Other columns can be one of the other types, but not ColTypeReference.
	Type ColumnType `json:"type"`
}

// EnumTable describes a table that contains enumerated values. The resulting Go code will be a type
// with constants for each value in the database. An EnumTable is processed with its data at compile time
// and cannot be modified by the application. The values can be stored in the database for data integrity purposes,
// and in case other processes are accessing the database, but otherwise the values are not accessed in the database.
type EnumTable struct {
	// Name is the name of the table in the database.
	// The name should have the Database.EnumTableSuffix value as a suffix.
	Name string `json:"name"`
	// For databases that support schemas, this is the name of the schema of the table.
	// Leave blank for the default schema.
	// Databases that do not support schemas will have this prepended to the name of the table.
	Schema string `json:"schema,omitempty"`

	// Fields describe the fields defined in the enum table.
	// The first field name MUST be the id field, and 2nd MUST be the name field.
	// The others are optional extra fields.
	Fields []*EnumField `json:"fields"`

	// Values are the enum values themselves.
	// Each entry must have the same number of items as are in the Fields slice and correspond to those types.
	// The first item in each entry should be an integer that will represent the value of the enumerated type, and the
	// value that the database will store.
	// The second item in each entry is a title string that will be the ToString result for that enumerated
	// value, and will also be used to create the constant name for the value.
	Values [][]interface{} `json:"values"`

	// Title is the name of the object when describing it to humans.
	// If creating a multi-language app, your app would provide translation from this string to the language of choice.
	// Can be multiple words. Should be lower-case. The app will use github.com/goradd/strings.ReverseTitle() to capitalize this if needed.
	// If left blank, the app will base this on the Name of the table.
	Title string `json:"title,omitempty"`

	// TitlePlural is the plural form of the Title.
	TitlePlural string `json:"title_plural,omitempty"`

	// Identifier is the corresponding Go object name. It must obey Go identifier labeling rules. Leave blank
	// to base it on the Name.
	Identifier string `json:"identifier,omitempty"`

	// IdentifierPlural is the plural form of Identifier.
	IdentifierPlural string `json:"identifier_plural,omitempty"`
}

func (t *EnumTable) QualifiedName() string {
	if t.Schema == "" {
		return t.Name
	} else {
		return t.Schema + "." + t.Name
	}
}

func (t *EnumTable) FillDefaults(suffix string) {
	name := strings.TrimSuffix(t.QualifiedName(), suffix)
	if t.Title == "" {
		t.Title = strings2.Title(name)
	}
	if t.TitlePlural == "" {
		t.TitlePlural = strings2.Plural(t.Title)
	}
	if t.Identifier == "" {
		t.Identifier = snaker.SnakeToCamelIdentifier(name)
	}
	if t.IdentifierPlural == "" {
		t.IdentifierPlural = strings2.Plural(t.Identifier)
	}
	for _, f := range t.Fields {
		if f.Title == "" {
			f.Title = strings2.Title(f.Name)
		}
		if f.TitlePlural == "" {
			f.TitlePlural = strings2.Plural(f.Title)
		}
		if f.Identifier == "" {
			f.Identifier = snaker.SnakeToCamelIdentifier(f.Name)
		}
		if f.IdentifierPlural == "" {
			f.IdentifierPlural = strings2.Plural(f.Identifier)
		}
	}
}

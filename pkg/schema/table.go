package schema

import (
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
)

// Table represents the metadata for a table in the database.
type Table struct {
	// Name is the name of the table or object as used in the database.
	// Must be unique within the table or schema if one is provided.
	// Should be lower_snake_case.
	Name string `json:"name"`

	// For databases that support schemas, this is the name of the schema of the table.
	// Leave blank for the default schema.
	// Databases that do not support schemas will have this prepended to the name of the table.
	Schema string `json:"schema,omitempty"`

	// Columns is a list of Column objects, one for each column in the table.
	Columns []*Column `json:"columns"`

	// MultiColumnIndexes will be used to generate multi-column getter functions.
	// In databases that support indexes, this will create a multi-column index in the database.
	// Single-column indexes are defined in the Column structure.
	MultiColumnIndexes []MultiColumnIndex `json:"multi_column_indexes,omitempty"`

	// Title is the name of the object when describing it to humans.
	// This is not used by the ORM, but may be used by UI generators.
	// If creating a multi-language app, your app would provide translation from this string to the language of choice.
	// Can be multiple words.
	// If left blank, the app will base this on the Name of the table.
	Title string `json:"title,omitempty"`

	// TitlePlural is the plural form of the Title.
	TitlePlural string `json:"title_plural,omitempty"`

	// Identifier is the corresponding Go object name. It must obey Go identifier labeling rules.
	Identifier string `json:"identifier,omitempty"`

	// IdentifierPlural is the plural form of Identifier.
	IdentifierPlural string `json:"identifier_plural,omitempty"`

	// Key is a value that helps synchronize changes to the table description.
	// It is assigned by the analyzer and should not be changed.
	Key string `json:"key"`

	// NoOrm will prevent the table from generating code or being used by the ORM.
	// You will still be able to access the table through direct calls to the database.
	// Not recommended for tables that are involved in any reference or association relationships between tables.
	NoOrm bool `json:"no_orm,omitempty"`

	// TODO: initial values
}

// QualifiedName returns the name to use to refer to the table
// in the database, including the schema if one is provided.
func (t *Table) QualifiedName() string {
	if t.Schema == "" {
		return t.Name
	} else {
		return t.Schema + "." + t.Name
	}
}

func (t *Table) FillDefaults(db *Database) {
	if t.Title == "" {
		t.Title = strings2.Title(t.Name)
	}
	if t.TitlePlural == "" {
		t.TitlePlural = strings2.Plural(t.Title)
	}
	if t.Identifier == "" {
		t.Identifier = snaker.SnakeToCamelIdentifier(t.QualifiedName())
	}
	if t.IdentifierPlural == "" {
		t.IdentifierPlural = strings2.Plural(t.Identifier)
	}

	for _, c := range t.Columns {
		c.FillDefaults(db, t)
	}
}

package schema

import (
	"fmt"
	strings2 "github.com/goradd/strings"
	"strings"
	"time"
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

	// WriteTimeout is used to wrap write transactions with a timeout on their contexts.
	// Leaving it as zero will use the Database.WriteTimeout value.
	// Use a duration format understood by time.ParseDuration.
	WriteTimeout string `json:"write_timeout,omitempty"`

	// ReadTimeout is used to wrap read transactions with a timeout on their contexts.
	// Leaving it as zero will use the Database.ReadTimeout value.
	// Use a duration format understood by time.ParseDuration.
	ReadTimeout string `json:"read_timeout,omitempty"`

	// NoTest indicates that the table should NOT have an automated test generated for it.
	NoTest bool `json:"no_test,omitempty"`

	// Columns is a list of Column objects, one for each column in the table.
	// This does not include columns associated with references.
	// Use AllColumns() to get the list of all columns, including the foreign key columns.
	Columns []*Column `json:"columns"`

	// References is a list of links to other tables, also known as foreign keys.
	References []*Reference `json:"references,omitempty"`

	// Indexes will be used to generate additional indexes and getter functions.
	// Single-column indexes can also be defined in Column and Reference.
	// You can use this to specify multiple types of indexes on the same column(s), including a
	// composite primary key. Note that some databases do not support composite primary keys and
	// will convert a composite primary key to a composite unique key, and then will automatically
	// generate a single unique primary key.
	// Composite primary keys with an auto generated primary key is not supported in the orm. To
	// implement this, you will need to manually generate or assign the primary key columns.
	Indexes []Index `json:"indexes,omitempty"`

	// Identifier is the corresponding Go object name.
	// It must obey Go identifier labeling rules.
	// Should be CamelCase.
	// If empty, will be based on Name.
	Identifier string `json:"identifier,omitempty"`

	// IdentifierPlural is the plural form of Identifier.
	// If blank, will be base on Identifier.
	IdentifierPlural string `json:"identifier_plural,omitempty"`

	// Label is the name of the object when describing it to humans.
	// If creating a multi-language app, your app would provide translation from this string to the language of choice.
	// Can be multiple words.
	// If left blank, will be base on Identifier.
	Label string `json:"label,omitempty"`

	// LabelPlural is the plural form of the Label.
	// If left blank, will be based on Label.
	LabelPlural string `json:"label_plural,omitempty"`

	// Comment is a place to put a comment in the json description file.
	// If the database driver supports it, it may be put in the database.
	Comment string `json:"comment,omitempty"`
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

// Clean fills in defaults and does some processing on the references and indexes.
func (t *Table) Clean(db *Database) error {
	if strings.ContainsRune(t.Name, '.') {
		return fmt.Errorf("table name %q cannot contain a period", t.Name)
	}
	if strings.ContainsRune(t.Schema, '.') {
		return fmt.Errorf("table schema %q cannot contain a period", t.Schema)
	}

	if t.WriteTimeout != "" {
		if _, err := time.ParseDuration(t.WriteTimeout); err != nil {
			return fmt.Errorf("invalid WriteTimeout %q: %w", t.WriteTimeout, err)
		}
	}
	if t.ReadTimeout != "" {
		if _, err := time.ParseDuration(t.ReadTimeout); err != nil {
			return fmt.Errorf("invalid ReadTimeout %q: %w", t.ReadTimeout, err)
		}
	}

	if len(t.Columns) == 0 && len(t.References) == 0 {
		return fmt.Errorf("table %q does not have any columns or references", t.QualifiedName())
	}

	for _, c := range t.Columns {
		if err := c.infer(db, t); err != nil {
			return err
		}
		if c.IndexLevel != IndexLevelNone {
			t.Indexes = append(t.Indexes, Index{IndexLevel: c.IndexLevel, Columns: []string{c.Name}})
		}
	}

	for _, ref := range t.References {
		if err := ref.infer(db, t); err != nil {
			return err
		}
		if ref.IndexLevel == IndexLevelNone {
			if !t.columnIsIndexed(ref.Column) {
				ref.IndexLevel = IndexLevelIndexed
			}
		}
		if ref.IndexLevel != IndexLevelNone {
			t.Indexes = append(t.Indexes, Index{IndexLevel: ref.IndexLevel, Columns: []string{ref.Column}})
		}
	}

	var hasPk bool
	for _, m := range t.Indexes {
		if m.IndexLevel == IndexLevelPrimaryKey {
			if hasPk {
				return fmt.Errorf("table %s cannot have multiple primary keys", t.QualifiedName())
			}
			hasPk = true
		}
	}
	if !hasPk {
		return fmt.Errorf("table %s has no primary key", t.QualifiedName())
	}
	return nil
}

func (t *Table) fillDefaults(db *Database) {
	if t.Identifier == "" {
		t.Identifier = strings2.SnakeToCamel(t.Name)
	}
	if t.IdentifierPlural == "" {
		t.IdentifierPlural = strings2.Plural(t.Identifier)
	}
	if t.Label == "" {
		t.Label = strings2.Title(t.Identifier)
	}
	if t.LabelPlural == "" {
		t.LabelPlural = strings2.Plural(t.Label)
	}
	for _, c := range t.Columns {
		c.fillDefaults()
	}
	for _, r := range t.References {
		r.fillDefaults(db, t)
	}
}

// PrimaryKeyColumns returns the names of the primary key columns of the table, or nil if not found.
// Note that these names may refer to reference columns.
// This only works after Clean has been called.
func (t *Table) PrimaryKeyColumns() []string {
	for _, i := range t.Indexes {
		if i.IndexLevel == IndexLevelPrimaryKey {
			return i.Columns
		}
	}
	return nil
}

// FindColumn returns the named column of the table, or nil if not found.
// This does not include columns in references.
func (t *Table) FindColumn(n string) *Column {
	for _, c := range t.Columns {
		if c.Name == n {
			return c
		}
	}
	return nil
}

func (t *Table) columnIsIndexed(n string) bool {
	for _, i := range t.Indexes {
		for _, c := range i.Columns {
			if c == n {
				return true
			}
		}
	}
	return false
}

package schema

import (
	"fmt"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"log/slog"
	"strings"
)

// AssociationReference describes a link from an association table.
type AssociationReference struct {
	// Table is the name of the table being referenced.
	// If Table is the same as the column's table, it creates a parent-child relationship.
	// Should match a Table.Name value in another table.
	Table string `json:"table"`

	// Schema is the schema that Table is in.
	Schema string `json:"schema,omitempty"`

	// Column is the name of the column created in the association table to hold a duplicate of the private key in Table.
	// It will default to a name based on Table and the name of the primary key column in Table.
	Column string `json:"column,omitempty"`
	/* Future expansion for composite keys:
	// Maps a column name in the association table to a column in the composite primary key in Table.
	// If empty, will Clean the values from the names of the columns in Table.
	Columns map[string]string
	*/

	// Identifier is the Go name used for the referenced object.
	// If not specified, will be based on Column.
	Identifier string `json:"identifier,omitempty"`

	// IdentifierPlural is the plural of Identifier.
	// Will be based on Identifier if left blank.
	IdentifierPlural string `json:"identifierPlural,omitempty"`

	// Label is the human-readable name for the referenced object.
	// If not specified, will be based on Identifier.
	Label string `json:"label,omitempty"`

	// LabelPlural is the plural of Label
	LabelPlural string `json:"labelPlural,omitempty"`
}

func (r *AssociationReference) QualifiedTableName() string {
	if r.Schema != "" {
		return fmt.Sprintf("%s.%s", r.Table, r.Schema)
	}
	return r.Table
}

// AssociationTable describes a table in the database that will be used to create a many-to-many
// relationship between two tables.
type AssociationTable struct {
	// Table is the name of the association table in the database. It should be lower_snake_case and end
	// with the Database.AssnTableSuffix value.
	// Example: "project_team_member_assn"
	Table string `json:"name"`
	// For databases that support schemas, this is the name of the schema of the table.
	// Leave blank for the default schema.
	// Databases that do not support schemas will have this prepended to the name of the table if this is present.
	Schema string `json:"schema,omitempty"`

	// Ref1 references the first table in the many-many relationship
	Ref1 AssociationReference `json:"ref1"`
	// Ref2 references the second table in the many-many relationship
	Ref2 AssociationReference `json:"ref2"`

	// Comment is a place to put a comment in the JSON schema file.
	// If the database driver supports it, it may be put in the database.
	Comment string `json:"comment,omitempty"`
}

func (t *AssociationTable) QualifiedTableName() string {
	if t.Schema == "" {
		return t.Table
	} else {
		return t.Schema + "." + t.Table
	}
}

func (t *AssociationTable) infer(db *Database) error {
	if err := t.inferRef(db, &t.Ref1); err != nil {
		return err
	}
	if err := t.inferRef(db, &t.Ref2); err != nil {
		return err
	}
	if t.Table == "" {
		slog.Error("Table name not specified in association table")
		return fmt.Errorf("table not specified in association table")
	}
	return nil
}

func (t *AssociationTable) inferRef(db *Database, ref *AssociationReference) error {
	if ref.Table == "" {
		slog.Error("Table in ref not specified in association table",
			slog.String("table", t.Table))
		return fmt.Errorf("table not specified in association table %s", t.Table)
	}
	if table := db.FindTable(ref.Table); table != nil {
		if ref.Column == "" {
			pks := table.PrimaryKeyColumns()
			if len(pks) == 1 {
				ref.Column = ref.Table + "_" + pks[0]
			} else {
				slog.Error("Table does not have a single primary key",
					slog.String("table", table.Name))
				return fmt.Errorf("table %s does not have a single primary key", table.Name)
			}
		}
	} else {
		slog.Error("A table referred to in an association table was not found in the schema",
			slog.String("table", ref.Table))
		return fmt.Errorf("An association table reference was not found")
	}
	return nil
}

// fillDefaults will fill default values where none have been set.
func (t *AssociationTable) fillDefaults(db *Database) {
	t.fillRefDefaults(db, &t.Ref1)
	t.fillRefDefaults(db, &t.Ref2)
}

func (t *AssociationTable) fillRefDefaults(db *Database, ref *AssociationReference) {
	if ref.Identifier == "" {
		table := db.FindTable(ref.Table)
		pks := table.PrimaryKeyColumns()
		if len(pks) == 1 {
			objName := strings.TrimSuffix(ref.Column, "_"+pks[0])
			ref.Identifier = snaker.SnakeToCamelIdentifier(objName)
		}
	}
	if ref.IdentifierPlural == "" {
		ref.IdentifierPlural = strings2.Plural(ref.Identifier)
	}
	if ref.Label == "" {
		ref.Label = strings2.Title(ref.Identifier)
	}
	if ref.LabelPlural == "" {
		ref.LabelPlural = strings2.Plural(ref.Label)
	}
}

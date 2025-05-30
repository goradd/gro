package schema

import (
	"fmt"
	strings2 "github.com/goradd/strings"
	"log/slog"
)

// AssociationTable describes a table in the database that will be used to create a many-to-many
// relationship between two tables.
type AssociationTable struct {
	// Name is the name of the association table in the database. It should be lower_snake_case and end
	// with the Database.AssnTableSuffix value. If empty, will be created using Name1 and Name2.
	// Example: "project_team_member_assn"
	Name string `json:"name,omitempty"`
	// For databases that support schemas, this is the name of the schema of the table.
	// Leave blank for the default schema.
	// Databases that do not support schemas will have this prepended to the name of the table if this is present.
	Schema string `json:"schema,omitempty"`

	// Table1 is the name of first table in the association.
	// If Schema is set, it will be expected to be in the same schema.
	// This table name must exist in the list of tables in the database, and must
	// not be an enum table.
	Table1 string `json:"table1"`

	// Name1 is the name prefix used to name the objects pointed to by Table1.
	// If empty, it will be based on Table1.
	// Should be lower_snake_case.
	// Example: "team_member".
	Name1 string `json:"name1,omitempty"`

	// Identifier1 is the singular Go name that will be used for one object pointed at.
	// If empty, will be based on Name1. Should be CamelCase.
	// Example: "TeamMember"
	Identifier1 string `json:"identifier1,omitempty"`

	// Identifier1Plural is the plural Go name that will be used for the objects pointed at.
	// Should be CamelCase.
	// If empty, will be based on Identifier1.
	// Example: "TeamMembers"
	Identifier1Plural string `json:"identifier1_plural,omitempty"`

	// Label1 is the singular description that will be used to describe the objects to human readers.
	// For example a "person" table might be used in a relationship
	// to describe members of a group. The label in that case could be "Team Member".
	// If empty, it will be based on Identifier1.
	Label1 string `json:"label1,omitempty"`

	// Label1Plural is the plural description that will be used to describe the objects to human readers.
	// This will be used to create the corresponding Go identifier.
	// Example: "Team Members"
	// The default will be based on Label1.
	Label1Plural string `json:"label1_plural,omitempty"`

	// Table2 is the name of second table in the association.
	Table2 string `json:"table2"`

	// The name prefix of the column in the association table that will be used to point to table 2.
	Name2 string `json:"name2,omitempty"`

	// Identifier2 is the singular Go name that will be used for one object pointed at.
	Identifier2 string `json:"identifier2,omitempty"`

	// Identifier2Plural is the plural Go name that will be used for the objects pointed at.
	Identifier2Plural string `json:"identifier2_plural,omitempty"`

	// Label2 is the singular description that will be used to describe the objects to human readers.
	Label2 string `json:"label2,omitempty"`

	// Label2Plural is the plural description that will be used to describe the objects to human readers.
	Label2Plural string `json:"label2_plural,omitempty"`

	// Comment is a place to put a comment in the JSON schema file.
	// If the database driver supports it, it may be put in the database.
	Comment string `json:"comment,omitempty"`

	// Column1 holds the inferred name of Column1.
	Column1 string
	// Column2 holds the inferred name of column2.
	Column2 string
}

func (t *AssociationTable) QualifiedName() string {
	if t.Schema == "" {
		return t.Name
	} else {
		return t.Schema + "." + t.Name
	}
}

func (t *AssociationTable) infer(db *Database, assn_suffix string) error {
	if t.Table1 == "" {
		slog.Error("Table1 not specified in association table",
			slog.String("table", t.Name))
		return fmt.Errorf("Table1 not specified in association table %s", t.Name)
	}
	if t.Table2 == "" {
		slog.Error("Table2 not specified in association table",
			slog.String("table", t.Name))
		return fmt.Errorf("Table2 not specified in association table %s", t.Name)
	}

	if t.Name1 == "" {
		t.Name1 = t.Table1
	}
	if t.Name2 == "" {
		t.Name2 = t.Table2
	}

	for _, table := range db.Tables {
		if table.Name == t.Table1 {
			t.Column1 = t.Name1 + "_" + table.PrimaryKeyColumn().Name
		}
		if table.Name == t.Table2 {
			t.Column2 = t.Name2 + "_" + table.PrimaryKeyColumn().Name
		}
	}
	if t.Column1 == "" {
		slog.Error("Table1 of an association table was not found in the schema",
			slog.String("table", t.Name))
		return fmt.Errorf("Table1 of association table %s was not found in the schema", t.Name)
	}
	if t.Column2 == "" {
		slog.Error("Table2 of an association table was not found in the schema",
			slog.String("table", t.Name))
		return fmt.Errorf("Table2 of association table %s was not found in the schema", t.Name)
	}
	return nil
}

// fillDefaults will fill default values where none have been set.
func (t *AssociationTable) fillDefaults(assn_suffix string) {
	if t.Identifier1 == "" {
		t.Identifier1 = strings2.SnakeToCamel(t.Name1)
	}
	if t.Identifier1Plural == "" {
		t.Identifier1Plural = strings2.Plural(t.Identifier1)
	}
	if t.Label1 == "" {
		t.Label1 = strings2.Title(t.Identifier1)
	}
	if t.Label1Plural == "" {
		t.Label1Plural = strings2.Plural(t.Label1)
	}

	if t.Identifier2 == "" {
		t.Identifier2 = strings2.SnakeToCamel(t.Label2)
	}
	if t.Identifier2Plural == "" {
		t.Identifier2Plural = strings2.Plural(t.Identifier2)
	}
	if t.Label2 == "" {
		t.Label2 = strings2.Title(t.Identifier2)
	}
	if t.Label2Plural == "" {
		t.Label2Plural = strings2.Plural(t.Label2)
	}

	if t.Name == "" {
		t.Name = t.Name1 + "_" + t.Name2 + "_" + assn_suffix
	}
	if t.Name1 == t.Name2 {
		slog.Warn("Name1 and Name2 cannot be the same. They have been modified.",
			slog.String("table", t.Name))
		t.Name1 = t.Name1 + "_1"
		t.Name2 = t.Name2 + "_2"
	}
}

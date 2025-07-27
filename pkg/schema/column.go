package schema

import (
	"fmt"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"golang.org/x/exp/slog"
)

// Column represents a database column with its attributes and associated metadata.
type Column struct {
	// Name is the name of the column in the database.
	//
	// If an enum column, by convention the name can be one of the following forms:
	//  - enumtype
	//	- tablename_enumtype
	//  - enumtype_enum
	//  - tablename_enumtype_enum
	// For example, if this table is named "person", and the enum table is named "status",
	// then by convention the name can be: status, person_status, status_enum, or person_status_enum.
	// If this convention is not followed, then the EnumTable should include the name of the enum table.
	Name string `json:"name"`

	// Type is the type of column. See the doc for ColumnType for more info.
	Type ColumnType `json:"type"`

	// SubType further describes how the database treats the type.
	SubType ColumnSubType `json:"sub_type,omitempty"`

	// If a string column, Size is the maximum length of runes that the column can accommodate.
	// If a []byte column, Size is the maximum number of bytes allowed in the column.
	// If an int, unsigned int, or float, Size is the number of bits allowed in the number and will also
	// determine the Go number type that will represent the column. This can be a zero in order to use the default,
	// which will be 32-bits for ints (the SQL default), or 64-bits for floats.
	Size uint64 `json:"size,omitempty"`

	// DefaultValue is the value that this field will be initialized to when a new object is created.
	// If not specified, it will be the zero value of the column's type.
	// Non-nullable columns that do not have a default value are required to be set by the application
	// before the object is saved.
	// Time columns can use the string "now" to set the value to the current time when the object
	// is first saved, and "update" to also set the value to the current time every time the object
	// is modified and saved.
	DefaultValue interface{} `json:"default_value,omitempty"`

	// IsNullable is true if the column can be given a NULL value.
	IsNullable bool `json:"nullable,omitempty"`

	// IndexLevel indicates what kind of single-column index is associated with this column.
	// ColTypeAutoPrimaryKey columns by default will be given a single primary key index.
	// At least one primary key index must be specified.
	// See Table.Indexes for specifying a multi-column index.
	IndexLevel IndexLevel `json:"index_level,omitempty"`

	// Identifier is the name of the column in Go code. Leave blank to base it on the Name.
	// Should be CamelCase. For example: "LastName".
	Identifier string `json:"identifier,omitempty"`

	// Label is the human-readable description of the item. Leave blank to base it on the Identifier.
	// Should be title case.
	// For example: "Last Name".
	Label string `json:"label,omitempty"`

	// DatabaseDefinition contains database specific extra information on the column that helps the database driver
	// recreate the column in the database if needed. The top key is a db.DriverType constant, and the secondary key is the
	// type of information. For example, a DECIMAL field might look like this:
	//  {"mysql":{"type":"decimal(5,2)"},"sqllite":{"type":"string"}}
	// This would indicate that in Mysql, the column is defined as DECIMAL(5,2), but in Sqllite, as a string.
	// The information recognized is specific to the database driver.
	DatabaseDefinition map[string]map[string]interface{} `json:"database_def,omitempty"`

	// Comment is a place to put a comment in the JSON description file. If the database driver supports it, it may be put in the database..
	Comment string `json:"comment,omitempty"`

	// EnumTable is the enum table if the Type is ColTypeEnum.
	EnumTable string `json:"enum_table,omitempty"`
}

// infer creates some required values if not specified and does some validity checks.
func (c *Column) infer(db *Database, table *Table) error {
	if c.Name == "" {
		slog.Error("Column name is empty",
			slog.String("table", table.Name))
		return fmt.Errorf("missing column name in table %s", table.Name)
	}
	if c.Type == ColTypeEnum && c.EnumTable == "" {
		// Infer the table from the name of the column
		for _, e := range db.EnumTables {
			if e.Name == c.Name ||
				e.Name == table.Name+"_"+c.Name ||
				e.Name == c.Name+db.EnumTableSuffix ||
				e.Name == table.Name+"_"+c.Name+db.EnumTableSuffix {
				c.EnumTable = e.Name
				break
			}
		}
		if c.EnumTable == "" {
			slog.Error("Enum table was not specified and could not be inferred",
				slog.String("table", table.Name),
				slog.String("column", c.Name))
			return fmt.Errorf("enum table was not specified and could not be inferred from table %s, column %s", table.Name, c.Name)
		}
	} else if c.Type == ColTypeAutoPrimaryKey {
		c.IndexLevel = IndexLevelPrimaryKey
	}
	return nil
}

func (c *Column) fillDefaults() {
	if c.Identifier == "" {
		c.Identifier = snaker.SnakeToCamelIdentifier(c.Name)
	}

	if c.Label == "" {
		c.Label = strings2.Title(c.Identifier)
	}

	if c.Size == 0 {
		if c.Type == ColTypeInt {
			c.Size = 32
		} else if c.Type == ColTypeFloat {
			c.Size = 64
		}
	}
}

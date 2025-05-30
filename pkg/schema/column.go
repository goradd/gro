package schema

import (
	"fmt"
	strings2 "github.com/goradd/strings"
	"golang.org/x/exp/slog"
	"strings"
)

// Column represents a database column with its attributes and associated metadata.
type Column struct {
	// Name is the name of the column in the database.
	//
	// If this is a reference column to another table, by convention this name should be
	// of the form "object_id", with "object" being the name of the object you want generated
	// and "id" being the name of the primary key column in the referenced table.
	// If specified in this way, then the values in Reference can be inferred. If not,
	// Reference will need to be set.
	//
	// If an enum column, by convention the name can be one of the following forms:
	//  - enumtype
	//	- tablename_enumtype
	//  - enumtype_enum
	//  - tablename_enumtype_enum
	// For example, if this table is named "person", and the enum table is named "status",
	// then by convention the name can be: status, person_status, status_enum, or person_status_enum.
	// If this convention is not followed, then the Reference should include the name of the enum table.
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
	IsNullable bool `json:"is_nullable,omitempty"`

	// IndexLevel indicates what kind of index is associated with this column.
	// ColTypeAutoPrimaryKey columns are automatically indexed uniquely, so this can be left blank for those columns.
	// This is for specifying single-column indexes.
	// See MultiColumnIndex for specifying a multi-column index or multi-column primary key.
	// Only one column in a table can have a single primary key column.
	IndexLevel IndexLevel `json:"index_level,omitempty"`

	// Reference can be set to specify additional information for references.
	// In particular, if the referenced table cannot be inferred from Name, then
	// it must be included here.
	Reference *Reference `json:"reference,omitempty"`

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
}

// infer creates some required values if not specified and does some validity checks.
func (c *Column) infer(db *Database, table *Table) error {
	if c.Name == "" {
		slog.Error("Column name is empty",
			slog.String("table", table.Name))
		return fmt.Errorf("missing column name in table %s", table.Name)
	}
	if c.Type == ColTypeEnum && c.Reference == nil {
		// Infer the table from the name of the column
		for _, e := range db.EnumTables {
			if e.Name == c.Name ||
				e.Name == table.Name+"_"+c.Name ||
				e.Name == c.Name+db.EnumTableSuffix ||
				e.Name == table.Name+"_"+c.Name+db.EnumTableSuffix {
				c.Reference = &Reference{
					Table: e.Name,
				}
				break
			}
		}
		if c.Reference == nil {
			slog.Error("Enum table was not specified and could not be inferred",
				slog.String("table", table.Name),
				slog.String("column", c.Name))
			return fmt.Errorf("enum table was not specified and could not be inferred from table %s, column %s", table.Name, c.Name)
		}
	} else if c.Type == ColTypeReference && c.Reference == nil {
		// try to infer referred table from column name
		parts := strings.Split(c.Name, "_")
		for i := 1; i < len(parts); i++ {
			tableName := strings.Join(parts[:i], "_")
			columnName := strings.Join(parts[i:], "_")
			if columnName == "" {
				break // not going to default to anything
			}
			for _, t := range db.Tables {
				if t.Name == tableName &&
					columnName == t.PrimaryKeyColumn().Name {
					c.Reference = &Reference{
						Table:  tableName,
						Column: columnName,
					}
				}
			}
		}
		if c.Reference == nil {
			slog.Error("Reference table was not specified and could not be inferred",
				slog.String("table", table.Name),
				slog.String("column", c.Name))
			return fmt.Errorf("referenced table was not specified and could not be inferred from table %s, column %s", table.Name, c.Name)
		}
	}
	return nil
}

func (c *Column) fillDefaults(db *Database, table *Table) {
	if c.Identifier == "" {
		objName := strings2.SnakeToCamel(c.Name)
		c.Identifier = SanitizeIdentifier(objName)
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

	if c.Reference != nil {
		if c.Reference.Table == "" {
			slog.Error("Reference table was not specified",
				slog.String("table", table.Name),
				slog.String("column", c.Name))
			return
		}

		if c.IndexLevel == IndexLevelNone {
			c.IndexLevel = IndexLevelIndexed
		}

		if c.Reference.Identifier == "" {
			c.Reference.Identifier = strings2.SnakeToCamel(c.Reference.Table)
		}
		if c.Reference.Label == "" {
			c.Reference.Label = strings2.Title(c.Reference.Identifier)
		}

		if c.Reference.ReverseIdentifier == "" {
			if c.Reference.Table+"_"+c.Reference.Column == c.Name {
				c.Reference.ReverseIdentifier = table.Identifier
			} else {
				c.Reference.ReverseIdentifier = c.Identifier + table.Identifier
			}
		}
		if c.Reference.ReverseIdentifierPlural == "" {
			c.Reference.ReverseIdentifierPlural = strings2.Plural(c.Reference.ReverseIdentifier)
		}

		if c.Reference.ReverseLabel == "" {
			c.Reference.ReverseLabel = strings2.Title(c.Reference.Identifier)
		}
		if c.Reference.ReverseLabelPlural == "" {
			c.Reference.ReverseLabelPlural = strings2.Plural(c.Reference.ReverseLabel)
		}
	}
}

func (c *Column) IsPrimaryKey() bool {
	return c.Type == ColTypeAutoPrimaryKey || c.IndexLevel == IndexLevelManualPrimaryKey
}

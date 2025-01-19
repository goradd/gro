package schema

import (
	. "github.com/goradd/all"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"slices"
	"strings"
)

// Column represents a database column with its attributes and associated metadata.
type Column struct {
	// Name is the name of the column in the database.
	// By convention, if this is a primary key, this value should be "id".
	// If this is a reference to another table, the name should end in "_id",
	// or whatever the Database.ReferenceSuffix value is set to.
	Name string `json:"name"`

	// Type is the type of column. See the doc for ColumnType for more info.
	// If this is an auto generated primary key column, or a reference, the type must be a string,
	// even if the underlying type in the database is a different type.
	Type ColumnType `json:"type"`

	// Identifier is the name of the column in Go code. Leave blank to base it on the Name.
	// References should keep the "ID" at the name to differentiate between the value of the
	// foreign key and the loaded object.
	// Enum references should NOT keep the "ID" value in this name.
	Identifier string `json:"identifier,omitempty"`

	// Title is the human-readable description of the item. Leave blank to base it on the Name.
	Title string `json:"title,omitempty"`

	// If a string column, Size is the maximum length of runes that the column can accommodate.
	// If a []byte column, Size is the maximum number of bytes allowed in the column.
	// If an int, unsigned int, or float, Size is the number of bits allowed in the number and will also
	// determine the Go number type that will represent the column. This can be a zero in order to use the default,
	// (int, uint or float64).
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
	IndexLevel IndexLevel `json:"index_level,omitempty"`

	// For string columns that have an index, this will cause the
	// index on the column to be sorted in a case-insensitive way and OrderBy and Unique tests to likewise be
	// case-insensitive.
	CaseInsensitive bool `json:"case_insensitive,omitempty"`

	// Reference is set when the column is a pointer to another table.
	// This is required for ColTypeReference, ColTypeEnum and ColTypeManyEnum tables.
	Reference *Reference `json:"reference,omitempty"`
}

// Reference is the additional information needed for reference type  and enum columns.
// For reference columns, if the IndexLevel of the containing column is Unique, it creates a one-to-one relationship.
// Otherwise, it is a one-to-many relationship.
type Reference struct {
	// If this column is a reference to an object in another table, this is the name of that other table.
	// If using schemas, the format should be "SchemaName.TableName".
	// The QueryName of the Column should end in "_id" or whatever the value of Database.ReferenceSuffix is for the database.
	// If Table is the same as the QueryName of the column's table, it creates a parent-child relationship.
	// This can point to an enum table, in which case Table should end in the EnumTableSuffix.
	Table string `json:"table"`

	// Identifier is the Go name used for the referenced object.
	Identifier string `json:"identifier,omitempty"`

	// Title is the human-readable name for the referenced object.
	Title string `json:"title,omitempty"`

	// The singular description of this table's objects as referred to by the referenced table.
	// If not specified, one will be created.
	ReverseTitle string `json:"reverse_title,omitempty"`

	// The plural description of this table's objects as referred to by the reference object.
	// If not specified, the ReverseTitle will be pluralized.
	ReverseTitlePlural string `json:"reverse_title_plural,omitempty"`

	// The singular Go identifier that will be used for the reverse relationships.
	// If not specified, the ReverseTitle will be used to create one.
	ReverseIdentifier string `json:"reverse_identifier,omitempty"`

	// The plural Go identifier that will be used for the reverse relationships.
	// If not specified, the ReverseIdentifier will be pluralized.
	ReverseIdentifierPlural string `json:"reverse_identifier_plural,omitempty"`
}

func (c *Column) FillDefaults(db *Database, table *Table) {
	if strings.HasSuffix(c.Name, db.EnumTableSuffix) {
		if c.Reference == nil {
			// Infer the table from the name of the column
			if slices.ContainsFunc(db.EnumTables, func(e *EnumTable) bool {
				return e.Name == c.Name
			}) {
				c.Reference = &Reference{
					Table: c.Name,
				}
			} else if slices.ContainsFunc(db.EnumTables, func(e *EnumTable) bool {
				return e.Name == table.Name+"_"+c.Name
			}) {
				c.Reference = &Reference{
					Table: table.Name + "_" + c.Name,
				}
			}
		}
		objName := strings.TrimSuffix(c.Name, db.EnumTableSuffix)
		if c.Title == "" {
			c.Title = strings2.Title(objName)
			if c.Type == ColTypeManyEnum {
				c.Title = strings2.Plural(c.Title)
			}
			if c.Identifier == "" {
				c.Identifier = snaker.ForceCamelIdentifier(c.Title)
			}
		}
	}
	if c.Title == "" {
		c.Title = strings2.Title(c.Name)
	}

	if c.Reference != nil {
		objName := strings.TrimSuffix(c.Name, db.ReferenceSuffix)

		if c.Reference.Identifier == "" {
			c.Reference.Identifier = snaker.ForceCamelIdentifier(objName)
		}
		if c.Reference.Title == "" {
			c.Reference.Title = strings2.Title(c.Reference.Identifier)
		}
		if objName == c.Reference.Table {
			objName = ""
		}
		if c.Reference.ReverseTitle == "" {
			c.Reference.ReverseTitle = If(objName, snaker.ForceCamelIdentifier(objName)+" "+table.Title, table.Title)
		}
		if c.Reference.ReverseTitlePlural == "" && c.IndexLevel != IndexLevelUnique {
			c.Reference.ReverseTitlePlural = strings2.Plural(c.Reference.ReverseTitle)
		}
		if c.Reference.ReverseIdentifier == "" {
			c.Reference.ReverseIdentifier = snaker.ForceCamelIdentifier(c.Reference.ReverseTitle)
		}
		if c.Reference.ReverseIdentifierPlural == "" && c.IndexLevel != IndexLevelUnique {
			c.Reference.ReverseIdentifierPlural = strings2.Plural(c.Reference.ReverseIdentifier)
		}

		// Enum references do not have a public difference between the name of the id field and the object itself.
		if c.Identifier == "" && strings.HasSuffix(c.Reference.Table, db.EnumTableSuffix) {
			c.Identifier = c.Reference.Identifier
		}
	}

	if c.Identifier == "" {
		c.Identifier = snaker.SnakeToCamelIdentifier(c.Name)
	}

}

func (c *Column) IsPrimaryKey() bool {
	return c.Type == ColTypeAutoPrimaryKey || c.IndexLevel == IndexLevelManualPrimaryKey
}

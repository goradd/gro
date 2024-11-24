package schema

import (
	any2 "github.com/goradd/any"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
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
	Type ColumnType `json:"type"`

	// Identifier is the name of the column in Go code. Leave blank to base it on the Name.
	Identifier string `json:"identifier"`

	// Title is the human-readable description of the item. Leave blank to base it on the Name.
	Title string `json:"title"`

	// If a string column, MaxLength is the maximum length of runes that the column can accommodate.
	// If a []byte column, MaxLength is the maximum number of bytes allowed in the column.
	// If an int, unsigned int, or float, MaxLength is the number of bits allowed in the number and will also
	// determine the Go number type that will represent the column. This can be a zero in order to use the default,
	// (int, uint or float64).
	MaxLength uint64 `json:"max_length"`

	// DefaultValue is the default value as specified by the database. We will initialize new ORM objects
	// with this value. Non-nullable values that do not have a default value are required to be set by the application.
	DefaultValue interface{} `json:"default_value"`

	// IsOptional is true if the column can be given a NULL value.
	IsOptional bool `json:"is_optional"`

	// IndexLevel indicates what kind of index is associated with this column.
	// ColTypeAutoPrimaryKey columns are automatically indexed uniquely, so this can be left blank for those columns.
	IndexLevel IndexLevel `json:"index_level"`

	// For string columns that have an index, this will cause the
	// index on the column to be sorted in a case-insensitive way and OrderBy and Unique tests to likewise be
	// case-insensitive.
	CaseInsensitive bool `json:"case_insensitive"`

	// Reference is required to be set if this is a reference type column.
	Reference *Reference `json:"reference,omitempty"`
}

// Reference is the additional information needed for reference type columns.
// If the IndexLevel of the containing column is Unique, it creates a one-to-one relationship.
// Otherwise, it is a one-to-many relationship.
type Reference struct {
	// If this column is a reference to an object in another table, this is the name of that other table.
	// If using schemas, the format should be "SchemaName.TableName".
	// The Name of the Column should end in "_id" or whatever the value of Database.ReferenceSuffix is for the database.
	// If the Table is the same as the Name of the column, it creates a parent-child relationship.
	// This can point to an enum table.
	Table string `json:"table"`

	// Column is the Name of the column referred to. Leave blank to refer to the primary key field
	// of the Table.
	Column string `json:"column"`

	// The singular description of this table's objects as referred to by the referenced table.
	// If not specified, one will be created.
	ReverseTitle string `json:"reverse_title"`

	// The plural description of this table's objects as referred to by the reference object.
	// If not specified, the ReverseTitle will be pluralized.
	ReverseTitlePlural string `json:"reverse_title_plural,omitempty"`

	// The singular Go identifier that will be used for the reverse relationships.
	// If not specified, the ReverseTitle will be used to create one.
	ReverseIdentifier string `json:"reverse_identifier"`

	// The plural Go identifier that will be used for the reverse relationships.
	// If not specified, the ReverseIdentifier will be pluralized.
	ReverseIdentifierPlural string `json:"reverse_identifier_plural,omitempty"`
}

func (c *Column) FillDefaults(table *Table, referenceSuffix string) {
	if c.Title == "" {
		c.Title = strings2.Title(c.Name)
	}
	if c.Identifier == "" {
		c.Identifier = snaker.SnakeToCamelIdentifier(c.Name)
	}
	if c.Reference != nil {
		objName := strings.TrimSuffix(c.Name, referenceSuffix)
		if objName == c.Reference.Table {
			objName = ""
		}

		if c.Reference.ReverseTitle == "" {
			c.Reference.ReverseTitle = any2.If(objName, objName+" "+table.Title, table.Title)
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
	}
}

func (c *Column) IsPrimaryKey() bool {
	return c.Type == ColTypeAutoPrimaryKey || c.IndexLevel == IndexLevelManualPrimaryKey
}

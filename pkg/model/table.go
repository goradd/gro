package model

import (
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"log/slog"
	"spekary/goradd/orm/pkg/schema"
)

type Table struct {
	// DbKey is the key used to find the database in the global database cluster
	DbKey string
	// QueryName is the name of the database table or object in the database.
	QueryName string
	// Title is the name of the object when describing it to the world. Should be lower case.
	Title string
	// TitlePlural is the plural name of the object.
	TitlePlural string
	// Identifier is the name of the struct type when referring to it in go code.
	Identifier string
	// IdentifierPlural is the name of a collection of these objects when referring to them in go code.
	IdentifierPlural string
	// DecapIdentifier is the same as Identifier, but with first letter lower case.
	DecapIdentifier string
	// Columns is a list of Columns, one for each column in the table.
	Columns []*Column
	// Indexes are the indexes defined on the table.
	Indexes []Index
	// Options are key-value pairs of values that can be used to customize how code generation is performed
	Options map[string]interface{}

	// The following items are filled in by the importSchema process

	// columnMap is an internal map of the columns by database name of the column
	columnMap map[string]*Column
	// ManyManyReferences describe the many-to-many references pointing to this table
	ManyManyReferences []*ManyManyReference
}

func (t *Table) PrimaryKeyColumn() *Column {
	if len(t.Columns) == 0 {
		return nil
	}
	if !t.Columns[0].IsPk {
		return nil
	}
	return t.Columns[0]
}

func (t *Table) PrimaryKeyGoType() string {
	return t.PrimaryKeyColumn().Type.GoType()
}

// ColumnByName returns a Column given the query name of the column,
// or nil if not found.
func (t *Table) ColumnByName(name string) *Column {
	return t.columnMap[name]
}

// DefaultHtmlID is the default id of corresponding form object when used in generated HTML.
func (t *Table) DefaultHtmlID() string {
	defaultID := snaker.CamelToSnake(t.Identifier)
	defaultID = strings2.SnakeToKebab(defaultID)
	return defaultID
}

// FileName is the base name of generated file names that correspond to this database table.
// Typically, Go files are lower case snake case by convention.
func (t *Table) FileName() string {
	s := snaker.CamelToSnake(t.Identifier)
	if strings2.EndsWith(s, "_test") {
		// Go will ignore files that end with _test. If we somehow create a filename like this,
		// we add an underscore to make sure it is still included in a build.
		s = s + "_"
	}
	return s
}

// HasGetterName returns true if the given name is in use by one of the getters.
// This is used for detecting naming conflicts. Will also return an error string
// to display if there is a conflict.
func (t *Table) HasGetterName(name string) (hasName bool, desc string) {
	for _, c := range t.Columns {
		if c.Identifier == name {
			return false, "conflicts with column " + c.Identifier
		}
		for _, rr := range c.ReverseReferences {
			if rr.Reference.ReverseIdentifier == name {
				return false, "conflicts with reverse reference singular name " + rr.Reference.ReverseIdentifier
			}
			if rr.Reference.ReverseIdentifierPlural == name {
				return false, "conflicts with reverse reference plural name " + rr.Reference.ReverseIdentifierPlural
			}
		}
	}

	for _, mm := range t.ManyManyReferences {
		if mm.Identifier == name {
			return false, "conflicts with many-many singular name " + mm.Identifier
		}
		if mm.IdentifierPlural == name {
			return false, "conflicts with many-many plural name " + mm.IdentifierPlural
		}
	}
	return false, ""
}

// newTable will import the table provided by tableSchema.
func newTable(dbKey string, tableSchema *schema.Table) *Table {
	queryName := strings2.If(tableSchema.Schema == "", tableSchema.Name, tableSchema.Schema+"."+tableSchema.Name)
	t := &Table{
		DbKey:            dbKey,
		QueryName:        queryName,
		Title:            tableSchema.Title,
		TitlePlural:      tableSchema.TitlePlural,
		Identifier:       tableSchema.Identifier,
		IdentifierPlural: tableSchema.IdentifierPlural,
		columnMap:        make(map[string]*Column),
	}

	t.DecapIdentifier = strings2.Decap(tableSchema.Identifier)

	if t.Identifier == t.IdentifierPlural {
		slog.Warn("Table skipped: table " + t.QueryName + " is using a plural name.")
		return nil
	}

	var pkCount int
	for _, schemaCol := range tableSchema.Columns {
		newCol := newColumn(schemaCol)
		t.Columns = append(t.Columns, newCol)
		t.columnMap[newCol.QueryName] = newCol
		if newCol.IsPk {
			pkCount++
			if pkCount > 1 {
				slog.Warn("Table " + t.QueryName + " has a multi-column primary key. A multi-column unique index with a single column primary key is prefered.")
			}
		}
		if schemaCol.IndexLevel != schema.IndexLevelNone {
			idx := Index{Columns: []*Column{newCol},
				IsUnique: schemaCol.IndexLevel != schema.IndexLevelIndexed}
			t.Indexes = append(t.Indexes, idx)
		}
	}

	for _, idx := range tableSchema.MultiColumnIndexes {
		var columns []*Column
		for _, name := range idx.Columns {
			col := t.ColumnByName(name)
			if col == nil {
				slog.Warn("Table skipped: cannot find column " + name + " of table " + t.QueryName + " in multi-column index")
				return nil
			}
			columns = append(columns, col)
		}
		t.Indexes = append(t.Indexes, Index{IsUnique: idx.IsUnique, Columns: columns})
	}
	return t
}

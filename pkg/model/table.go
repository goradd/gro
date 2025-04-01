package model

import (
	"fmt"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"log/slog"
	"slices"
)

type Table struct {
	// DbKey is the key used to find the database in the global database cluster
	DbKey string
	// QueryName is the name of the database table or object in the database.
	QueryName string
	// Label is the name of the object when describing it to the world. Should be lower case.
	Label string
	// LabelPlural is the plural name of the object.
	LabelPlural string
	// Identifier is the name of the struct type when referring to it in go code.
	Identifier string
	// IdentifierPlural is the name of a collection of these objects when referring to them in go code.
	IdentifierPlural string
	// DecapIdentifier is the same as Identifier, but with first letter lower case.
	DecapIdentifier string
	// Columns is a list of Columns, one for each column in the table.
	// Primary keys are sorted to the front.
	Columns []*Column
	// Indexes are the indexes defined on the table.
	Indexes []Index
	// Options are key-value pairs of values that can be used to customize how code generation is performed
	Options map[string]interface{}
	// columnMap is an internal map of the columns by query name of the column
	columnMap map[string]*Column
	// ReverseReferences are the columns from other tables, or even this table,
	// that point to this column.
	ReverseReferences []*Column
	// ManyManyReferences describe the many-to-many references pointing to this table
	ManyManyReferences []*ManyManyReference
	// The cached optimistic locking column, if one is present
	lockColumn *Column
}

// PrimaryKeyColumn returns the primary key column
func (t *Table) PrimaryKeyColumn() *Column {
	if len(t.Columns) == 0 {
		return nil
	}
	return t.Columns[0]
}

func (t *Table) PrimaryKeyGoType() string {
	return t.PrimaryKeyColumn().GoType()
}

// ColumnByName returns a Column given the query name of the column,
// or nil if not found.
func (t *Table) ColumnByName(name string) *Column {
	return t.columnMap[name]
}

func (t *Table) VariableNamePlural() string {
	return LowerCaseIdentifier(t.IdentifierPlural)
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
		for _, rr := range t.ReverseReferences {
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

// HasAutoPK returns true if the table has an automatically generated primary key
func (t *Table) HasAutoPK() bool {
	return t.PrimaryKeyColumn().IsAutoPK
}

// SettableColumns returns an array of columns that are settable
func (t *Table) SettableColumns() []*Column {
	var out []*Column
	for _, c := range t.Columns {
		if c.HasSetter() {
			out = append(out, c)
		}
	}
	return out
}

// LockColumn returns the special column that will be used to implement optimistic locking, if one exists for the table.
func (t *Table) LockColumn() *Column {
	return t.lockColumn
}

// newTable will import the table provided by tableSchema.
// If an error occurs, nil is returned.
func newTable(dbKey string, tableSchema *schema.Table) *Table {
	queryName := strings2.If(tableSchema.Schema == "", tableSchema.Name, tableSchema.Schema+"."+tableSchema.Name)
	t := &Table{
		DbKey:            dbKey,
		QueryName:        queryName,
		Label:            tableSchema.Label,
		LabelPlural:      tableSchema.LabelPlural,
		Identifier:       tableSchema.Identifier,
		IdentifierPlural: tableSchema.IdentifierPlural,
		columnMap:        make(map[string]*Column),
	}

	t.DecapIdentifier = strings2.Decap(tableSchema.Identifier)

	if t.Identifier == t.IdentifierPlural {
		slog.Error("Table " + t.QueryName + " is using a plural name.")
		return nil
	}

	if len(tableSchema.Columns) == 0 {
		slog.Error("Table " + t.QueryName + " has no columns.")
		return nil
	}

	var pkCount int
	for _, schemaCol := range tableSchema.Columns {
		newCol := newColumn(schemaCol)
		newCol.Table = t
		if (newCol.SchemaSubType == schema.ColSubTypeTimestamp ||
			newCol.SchemaSubType == schema.ColSubTypeLock) && newCol.IsNullable {
			slog.Warn(fmt.Sprintf("Column %s:%s should not be nullable. Nullable status will be ignored.", t.QueryName, newCol.QueryName))
			newCol.IsNullable = false
		}

		if newCol.IsPrimaryKey {
			pkCount++
			if pkCount > 1 {
				slog.Error("Table " + t.QueryName + " has a multi-column primary key. Instead combine a multi-column unique index with a single column auto-generated primary key.")
				return nil
			}
			t.Columns = slices.Insert(t.Columns, 0, newCol)
		} else {
			t.Columns = append(t.Columns, newCol)
		}
		t.columnMap[newCol.QueryName] = newCol
		if schemaCol.IndexLevel != schema.IndexLevelNone {
			idx := Index{Columns: []*Column{newCol},
				IsUnique: schemaCol.IndexLevel != schema.IndexLevelIndexed}
			t.Indexes = append(t.Indexes, idx)
		}
		if schemaCol.SubType == schema.ColSubTypeLock {
			t.lockColumn = newCol
		}
	}

	for _, idx := range tableSchema.MultiColumnIndexes {
		var columns []*Column
		for _, name := range idx.Columns {
			col := t.ColumnByName(name)
			if col == nil {
				slog.Error("Cannot find column " + name + " of table " + t.QueryName + " in multi-column index")
				return nil
			}
			columns = append(columns, col)
		}
		t.Indexes = append(t.Indexes, Index{IsUnique: idx.IsUnique, Columns: columns})
	}
	return t
}

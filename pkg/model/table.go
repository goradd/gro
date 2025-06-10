package model

import (
	"fmt"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"log/slog"
	"strings"
	"time"
)

type Table struct {
	// DbKey is the key used to find the database in the global database cluster
	DbKey string
	// WriteTimeout is used to wrap write transactions with a timeout on their contexts.
	// Leaving it as zero will use the timeout in the database, if one is set.
	WriteTimeout time.Duration
	// ReadTimeout is used to protect read transactions with a timeout on their contexts.
	// Leaving it as zero will use the timeout in the database, if one is set.
	ReadTimeout time.Duration
	// NoTest indicates that the table should NOT have an automated test generated for it.
	NoTest bool
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
	// DecapIdentifier is the same as Identifier, but the first letter is lower case.
	DecapIdentifier string
	// Columns is a list of Columns, one for each column in the table, but does not include columns in references.
	Columns []*Column
	// Indexes are all the indexes defined on the table, single and multi-column, but not primary key.
	Indexes []Index
	// Options are key-value pairs of values that can be used to customize how code generation is performed
	Options map[string]interface{}
	// References are all the foreign keys. References contains additional columns.
	References []*Reference
	// ReverseReferences are the columns from other tables, or even this table,
	// that point to this table.
	ReverseReferences []*Reference
	// ManyManyReferences describe the many-to-many references pointing to this table
	ManyManyReferences []*ManyManyReference
	// The cached optimistic locking column, if one is present
	lockColumn *Column
	// columnMap is an internal map of the columns by query name of the column
	columnMap map[string]*Column
	// primaryKeyColumns is a cache of the primary key columns, which can include reference columns
	primaryKeyColumns []*Column
}

// PrimaryKeyColumn returns a single primary key column if the table is keyed on one column.
// If not, nil is returned.
func (t *Table) PrimaryKeyColumn() *Column {
	if len(t.primaryKeyColumns) != 1 {
		return nil
	}
	return t.primaryKeyColumns[0]
}

// PrimaryKeyColumns returns a slice of the primary key columns in the table.
// Note that some of these columns may be part of a reference.
func (t *Table) PrimaryKeyColumns() []*Column {
	return t.primaryKeyColumns
}

// PrimaryKeyGoTypes returns a comma separated list of types for the primary keys, such that it can be
// used in a return value list.
func (t *Table) PrimaryKeyGoTypes() string {
	var s []string
	for _, column := range t.PrimaryKeyColumns() {
		s = append(s, column.GoType())
	}
	return strings.Join(s, ", ")
}

// ColumnByName returns a Column given the query name of the column,
// or nil if not found.
func (t *Table) ColumnByName(name string) *Column {
	return t.columnMap[name]
}

func (t *Table) VariableNamePlural() string {
	return LowerCaseIdentifier(t.IdentifierPlural)
}

func (t *Table) WriteTimeoutConst() string {
	return durationConst(t.WriteTimeout)
}
func (t *Table) ReadTimeoutConst() string {
	return durationConst(t.ReadTimeout)
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
	}
	for _, ref := range t.References {
		if ref.Identifier == name {
			return false, "conflicts with reference " + ref.Identifier
		}
		if ref.Column.Identifier == name {
			return false, "conflicts with reference column " + ref.Column.Identifier
		}
	}

	for _, rr := range t.ReverseReferences {
		if rr.ReverseIdentifier == name {
			return false, "conflicts with reverse reference singular name " + rr.ReverseIdentifier
		}
		if rr.ReverseIdentifierPlural == name {
			return false, "conflicts with reverse reference plural name " + rr.ReverseIdentifierPlural
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
	pk := t.PrimaryKeyColumn()

	return pk != nil && pk.SchemaType == schema.ColTypeAutoPrimaryKey
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

// HasUniqueIndexes returns true if the table has at least one unique index.
func (t *Table) HasUniqueIndexes() bool {
	for _, idx := range t.Indexes {
		if idx.IsUnique {
			return true
		}
	}
	return false
}

// HasReferences returns true if the table has at least one forward reference.
func (t *Table) HasReferences() bool {
	return len(t.References) > 0
}

// HasReverseReferences returns true if the table has at least one reverse reference.
func (t *Table) HasReverseReferences() bool {
	return len(t.ReverseReferences) > 0
}

// HasManyManyReferences returns true if the table has at least one many-many reference.
func (t *Table) HasManyManyReferences() bool {
	return len(t.ManyManyReferences) > 0
}

// importTable will import the table provided by tableSchema.
// If an error occurs, it is logged and nil is returned.
// There are a number of dependencies here, so code order is important.
func (m *Database) importTable(tableSchema *schema.Table,
	writeTimeout time.Duration,
	readTimeout time.Duration,
) {

	// Override database wide timeouts with local values
	if tableSchema.WriteTimeout != "" {
		writeTimeout, _ = time.ParseDuration(tableSchema.WriteTimeout)
	}
	if tableSchema.ReadTimeout != "" {
		readTimeout, _ = time.ParseDuration(tableSchema.ReadTimeout)
	}

	t := &Table{
		DbKey:            m.Key,
		WriteTimeout:     writeTimeout,
		ReadTimeout:      readTimeout,
		QueryName:        tableSchema.QualifiedName(),
		Label:            tableSchema.Label,
		LabelPlural:      tableSchema.LabelPlural,
		Identifier:       tableSchema.Identifier,
		IdentifierPlural: tableSchema.IdentifierPlural,
		NoTest:           tableSchema.NoTest,
		columnMap:        make(map[string]*Column),
	}

	t.DecapIdentifier = strings2.Decap(tableSchema.Identifier)

	if t.Identifier == t.IdentifierPlural {
		slog.Error("Table skipped. Table identifier is plural word.",
			slog.String(db.LogTable, t.QueryName))
		return
	}

	for _, schemaCol := range tableSchema.Columns {
		newCol := m.importColumn(schemaCol)
		newCol.Table = t
		t.Columns = append(t.Columns, newCol)
		if _, ok := t.columnMap[newCol.QueryName]; ok {
			slog.Error("Table skipped. Table has two columns with the same name.",
				slog.String(db.LogTable, t.QueryName),
				slog.String(db.LogColumn, newCol.QueryName),
			)
			return
		}
		t.columnMap[newCol.QueryName] = newCol
		if schemaCol.SubType == schema.ColSubTypeLock {
			t.lockColumn = newCol
		}
	}

	// The following relies on the order of tables being processed
	// such that the referenced tables exist with primary keys.
	for _, schemaRef := range tableSchema.References {
		ref := m.importReference(schemaRef)
		if _, ok := t.columnMap[ref.Column.QueryName]; ok {
			slog.Error("Table skipped. Reference column name already exists in the table.",
				slog.String(db.LogTable, t.QueryName),
				slog.String(db.LogColumn, ref.Column.QueryName),
			)
			return
		}
		t.columnMap[ref.Column.QueryName] = ref.Column
		t.References = append(t.References, ref)
	}

	for _, idx := range tableSchema.Indexes {
		var columns []*Column
		for _, name := range idx.Columns {
			col := t.ColumnByName(name)
			if col == nil {
				slog.Error("Cannot find column in index",
					slog.String(db.LogTable, t.QueryName),
					slog.String(db.LogColumn, name))
				continue
			} else {
				columns = append(columns, col)
			}
		}
		if len(columns) > 0 {
			if idx.IndexLevel == schema.IndexLevelPrimaryKey {
				t.primaryKeyColumns = columns
			} else {
				t.Indexes = append(t.Indexes, Index{IsUnique: idx.IndexLevel == schema.IndexLevelUnique, Columns: columns})
			}
		}
	}
	m.Tables[t.QueryName] = t
}

func durationConst(d time.Duration) string {
	if d == 0 {
		return ""
	}
	// try from largest to smallest
	for _, u := range []struct {
		unit time.Duration
		name string
	}{
		{time.Hour, "time.Hour"},
		{time.Minute, "time.Minute"},
		{time.Second, "time.Second"},
		{time.Millisecond, "time.Millisecond"},
		{time.Microsecond, "time.Microsecond"},
		{time.Nanosecond, "time.Nanosecond"},
	} {
		if d%u.unit == 0 {
			return fmt.Sprintf("%d * %s", d/u.unit, u.name)
		}
	}
	// fallback to nanoseconds (always valid Go)
	return fmt.Sprintf("%d * time.Nanosecond", d)
}

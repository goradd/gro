package model

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/goradd/gro/db"
	"github.com/goradd/gro/schema"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
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
	// QueryName is the database's identifier for the table.
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
	LockColumn *Column
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

// PrimaryKeyType returns a type name for the primary key.
// If the primary key is a single column, this will be a go type.
// If a composite key, this will be the name of a struct type that should be generated to house
// the primary key values.
func (t *Table) PrimaryKeyType() string {
	if len(t.primaryKeyColumns) == 1 {
		return t.primaryKeyColumns[0].Type
	} else {
		return t.Identifier + "PrimaryKey"
	}
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
		if ref.ForeignKey.Identifier == name {
			return false, "conflicts with reference column " + ref.ForeignKey.Identifier
		}
	}

	for _, rr := range t.ReverseReferences {
		if rr.ReverseIdentifier == name {
			return false, "conflicts with reverse reference name " + rr.ReverseIdentifier
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

// AllColumns returns all the columns in the table, including foreign keys
func (t *Table) AllColumns() (out []*Column) {
	out = append(out, t.Columns...)
	for _, ref := range t.References {
		out = append(out, ref.ForeignKey)
	}
	return
}

// SettableColumns returns an array of columns that are settable, including foreign keys.
func (t *Table) SettableColumns() (out []*Column) {
	for _, c := range t.Columns {
		if c.HasSetter() {
			out = append(out, c)
		}
	}
	for _, ref := range t.References {
		out = append(out, ref.ForeignKey)
	}
	return out
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

// LockColumnQueryName returns the Column.QueryName of the lock column, or an empty string if there is no lock column.
func (t *Table) LockColumnQueryName() string {
	if t.LockColumn == nil {
		return ""
	} else {
		return t.LockColumn.QueryName
	}
}

// LockColumnIdentifier returns the Column.Identifier of the lock column, or an empty string if there is no lock column.
func (t *Table) LockColumnIdentifier() string {
	if t.LockColumn == nil {
		return ""
	} else {
		return t.LockColumn.Identifier
	}
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
			t.LockColumn = newCol
		}
	}

	var selfRefs []*schema.Reference

	// The following relies on the order of tables being processed
	// such that the referenced tables exist with primary keys.
	for _, schemaRef := range tableSchema.References {
		if schemaRef.Table == t.QueryName {
			// Handle self references after primary key indexes are processed
			selfRefs = append(selfRefs, schemaRef)
			continue
		}

		ref := m.importReference(t, schemaRef)
		if ref == nil {
			return
		}
		if _, ok := t.columnMap[ref.ForeignKey.QueryName]; ok {
			slog.Error("Table skipped. Reference column name already exists in the table.",
				slog.String(db.LogTable, t.QueryName),
				slog.String(db.LogColumn, ref.ForeignKey.QueryName),
			)
			return
		}
		t.columnMap[ref.ForeignKey.QueryName] = ref.ForeignKey
		t.References = append(t.References, ref)
	}

	// Process just the primary keys so that self references can see the primary key
	for _, idx := range tableSchema.Indexes {
		if idx.IndexLevel == schema.IndexLevelPrimaryKey {
			var columns []*Column
			for _, name := range idx.Columns {
				col := t.ColumnByName(name)
				if col == nil {
					slog.Error("Cannot find primary key column",
						slog.String(db.LogTable, t.QueryName),
						slog.String(db.LogColumn, name))
					continue
				} else {
					columns = append(columns, col)
				}
			}
			t.primaryKeyColumns = columns
		}
	}

	// Process self references
	for _, schemaRef := range selfRefs {
		ref := m.importReference(t, schemaRef)
		if ref == nil {
			return
		}
		if _, ok := t.columnMap[ref.ForeignKey.QueryName]; ok {
			slog.Error("Table skipped. Reference column name already exists in the table.",
				slog.String(db.LogTable, t.QueryName),
				slog.String(db.LogColumn, ref.ForeignKey.QueryName),
			)
			return
		}
		t.columnMap[ref.ForeignKey.QueryName] = ref.ForeignKey
		t.References = append(t.References, ref)
	}

	// Process the rest of the indexes
	for _, idx := range tableSchema.Indexes {
		if idx.IndexLevel != schema.IndexLevelPrimaryKey {
			var columns []*Column
			for _, name := range idx.Columns {
				col := t.ColumnByName(name)
				if col == nil {
					slog.Error("Cannot find primary key column",
						slog.String(db.LogTable, t.QueryName),
						slog.String(db.LogColumn, name))
					continue
				} else {
					columns = append(columns, col)
				}
			}
			t.Indexes = append(t.Indexes,
				Index{
					IsUnique:   idx.IndexLevel == schema.IndexLevelUnique,
					Columns:    columns,
					Identifier: idx.Identifier,
					Name:       idx.Name,
				})
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

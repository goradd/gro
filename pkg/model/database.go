package model

import (
	"cmp"
	"github.com/goradd/anyutil"
	"github.com/goradd/maps"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
	"log/slog"
	maps2 "maps"
	"slices"
	"strings"
	"time"
)

func FromSchema(s *schema.Database) *Database {
	if err := s.Clean(); err != nil {
		panic(err)
	}
	s.FillDefaults()
	var timeout time.Duration
	if s.WriteTimeout != "" {
		var err error
		timeout, err = time.ParseDuration(s.WriteTimeout)
		if err != nil {
			slog.Warn("invalid timeout",
				slog.Any(db.LogError, err))
			timeout = 0
		}
	}
	d := Database{
		Key:             s.Key,
		WriteTimeout:    timeout,
		EnumTableSuffix: s.EnumTableSuffix,
		AssnTableSuffix: s.AssnTableSuffix,
	}
	d.importSchema(s)
	return &d
}

// Database is the top level struct that contains a description of a database modeled as objects.
// It is used in code generation and query creation.
type Database struct {
	// The database key corresponding to its key in the global database cluster
	Key string
	// WriteTimeout is used to wrap write transactions with a timeout on their contexts.
	// Leaving it as zero will turn off timeout wrapping, allowing you to wrap individual calls as you
	// see fit.
	WriteTimeout time.Duration
	// ReadTimeout is used to wrap read transactions with a timeout on their contexts.
	// Leaving it as zero will turn off timeout wrapping, allowing you to wrap individual calls as you
	// see fit.
	ReadTimeout time.Duration
	// Tables are the tables in the database, keyed by database table name
	Tables map[string]*Table
	// Enums contains a description of the enumerated types linked to the database, keyed by database table name
	Enums map[string]*Enum

	// EnumTableSuffix is the text to string off the end of an enum table when converting it to a type name.
	// Defaults to "_enum".
	EnumTableSuffix string
	// Defaults to _assn
	AssnTableSuffix string
}

// importSchema will convert a database description to a model which generally treats
// tables as objects and columns as member variables.
func (m *Database) importSchema(schema *schema.Database) {
	m.Enums = make(map[string]*Enum)
	m.Tables = make(map[string]*Table)

	// deal with enum tables first
	for _, et := range schema.EnumTables {
		tt := newEnumTable(m.Key, et)
		m.Enums[tt.QueryName] = tt
	}

	// get the regular tables
	for _, table := range schema.Tables {
		t := newTable(m.Key, table)
		if t != nil {
			if t.WriteTimeout == 0 {
				t.WriteTimeout = m.WriteTimeout
			}
			if t.ReadTimeout == 0 {
				t.ReadTimeout = m.ReadTimeout
			}

			m.Tables[t.QueryName] = t
		}
	}

	// import references after the columns are in place
	for _, table := range schema.Tables {
		if t := m.Table(table.Name); t != nil {
			m.importReferences(t, table)
		}
	}

	for _, assn := range schema.AssociationTables {
		m.importAssociation(assn)
	}
}

// Analyzes an association table and creates special virtual columns in the corresponding tables it points to.
// Association tables are used by SQL databases to create many-many relationships. NoSQL databases can define their
// association columns directly and store an array of records on either end of the association.
func (m *Database) importAssociation(schemaAssn *schema.AssociationTable) {
	t1 := m.Table(schemaAssn.Table1)
	if t1 == nil {
		slog.Error("Missing associated table from association "+schemaAssn.Name,
			slog.String(db.LogTable, schemaAssn.Table1))
		return
	}
	t2 := m.Table(schemaAssn.Table2)
	if t2 == nil {
		slog.Error("Missing associated table from association "+schemaAssn.Table2,
			slog.String(db.LogTable, schemaAssn.Table2))
		return
	}

	if schemaAssn.Name1 == schemaAssn.Name2 {
		slog.Error("Name1 and Name2 cannot be the same",
			slog.String(db.LogTable, schemaAssn.Name))
		return
	}

	ref1 := makeManyManyRef(schemaAssn.Name, schemaAssn.Column1, schemaAssn.Column2, t1, t2, schemaAssn.Label2, schemaAssn.Label2Plural, schemaAssn.Identifier2, schemaAssn.Identifier2Plural)
	ref2 := makeManyManyRef(schemaAssn.Name, schemaAssn.Column2, schemaAssn.Column1, t2, t1, schemaAssn.Table1, schemaAssn.Label1Plural, schemaAssn.Identifier1, schemaAssn.Identifier1Plural)
	ref1.MM = ref2
	ref2.MM = ref1
}

func (m *Database) importReferences(t *Table, schemaTable *schema.Table) {
	for _, col := range schemaTable.Columns {
		m.importReference(t, col)
	}
}

func (m *Database) importReference(table *Table, schemaCol *schema.Column) {
	if schemaCol.Reference == nil {
		return
	}

	var thisCol *Column
	thisCol = table.ColumnByName(schemaCol.Name)
	refTable := m.Table(schemaCol.Reference.Table)
	et := m.Enum(schemaCol.Reference.Table)
	if refTable == nil && et == nil {
		slog.Error("Reference skipped, referenced table not found.",
			slog.String(db.LogTable, table.QueryName),
			slog.String(db.LogColumn, schemaCol.Name))
		return
	}
	if et != nil {
		if thisCol == nil {
			slog.Error("Could not find enum column",
				slog.String("table", table.QueryName),
				slog.String("column", schemaCol.Name))
			return
		}
		thisCol.EnumTable = et
		return
	}

	ref := &Reference{
		Table:                   refTable,
		Identifier:              schemaCol.Identifier,
		Label:                   schemaCol.Label,
		ReverseLabel:            schemaCol.Reference.ReverseLabel,
		ReverseLabelPlural:      schemaCol.Reference.ReverseLabelPlural,
		ReverseIdentifier:       schemaCol.Reference.ReverseIdentifier,
		ReverseIdentifierPlural: schemaCol.Reference.ReverseIdentifierPlural,
		DecapIdentifier:         strings2.Decap(schemaCol.Identifier),
		Column:                  thisCol,
	}

	// Fix up receiver type to match the primary key type of the refTable.
	// In autoId tables, this is always a string, but for manually entered primary keys, it could be anything.
	thisCol.ReceiverType = refTable.PrimaryKeyColumn().ReceiverType
	refTable.ReverseReferences = append(refTable.ReverseReferences, ref)
	table.References = append(table.References, ref)

}

// Table returns a Table from the database given the table name.
func (m *Database) Table(name string) *Table {
	if name == "" {
		return nil
	}
	if v, ok := m.Tables[name]; ok {
		return v
	} else {
		return nil
	}
}

// Enum returns an Enum from the database given the table name.
func (m *Database) Enum(name string) *Enum {
	if name == "" {
		return nil
	}
	return m.Enums[name]
}

// IsEnumTable returns true if the given name is the name of an enum table in the database
func (m *Database) IsEnumTable(name string) bool {
	_, ok := m.Enums[name]
	return ok
}

func isReservedIdentifier(s string) bool {
	switch s {
	case "break":
		return true
	case "case":
		return true
	case "chan":
		return true
	case "const":
		return true
	case "continue":
		return true
	case "default":
		return true
	case "defer":
		return true
	case "else":
		return true
	case "fallthrough":
		return true
	case "for":
		return true
	case "func":
		return true
	case "go":
		return true
	case "goto":
		return true
	case "if":
		return true
	case "import":
		return true
	case "interface":
		return true
	case "map":
		return true
	case "package":
		return true
	case "range":
		return true
	case "return":
		return true
	case "select":
		return true
	case "struct":
		return true
	case "switch":
		return true
	case "type":
		return true
	case "var":
		return true
	}
	return false
}

func LowerCaseIdentifier(s string) (i string) {
	if strings.Contains(s, "_") {
		i = snaker.ForceLowerCamelIdentifier(snaker.SnakeToCamelIdentifier(s))
	} else {
		// Not a snake string, but still might need some fixing up
		i = snaker.ForceLowerCamelIdentifier(s)
	}
	i = strings.TrimSpace(i)
	if isReservedIdentifier(i) {
		panic("Cannot use '" + i + "' as an identifier.")
	}
	if i == "" {
		panic("Cannot use blank as an identifier.")
	}
	return
}

func UpperCaseIdentifier(s string) (i string) {
	if strings.Contains(s, "_") {
		i = snaker.ForceCamelIdentifier(snaker.SnakeToCamelIdentifier(s))
	} else {
		// Not a snake string, but still might need some fixing up
		i = snaker.ForceCamelIdentifier(s)
	}
	i = strings.TrimSpace(i)
	if i == "" {
		panic("Cannot use blank as an identifier.")
	}
	return
}

// MarshalOrder returns an array of tables in the order they should be marshaled such that
// if they then get unmarshalled in the same order, there will not be problems with foreign
// keys not existing when an object is eventually saved.
// Note that it cannot do this for circular references, and so if your database has circular
// references, including self references, any foreign key checking will need to be turned off while importing the database.
func (m *Database) MarshalOrder() (tables []*Table) {
	var unusedTables maps.SliceSet[*Table]

	unusedTables.SetSortFunc(func(a, b *Table) bool {
		return a.QueryName < b.QueryName
	})
	unusedTables.Add(slices.Collect(maps2.Values(m.Tables))...)
	// First add the tables that have no forward references
	for { // repeat until unusedTables is empty
		var newTables []*Table
	nexttable:
		for t := range unusedTables.All() {
			for _, ref := range t.References {
				// skip this table if it has references to a table we have not yet seen
				if !slices.Contains(tables, ref.Table) &&
					!slices.Contains(newTables, ref.Table) &&
					t != ref.Table {
					continue nexttable
				}

			}
			// This has no forward references we care about
			newTables = append(newTables, t)
		}
		slices.SortFunc(newTables, func(a, b *Table) int {
			if a.QueryName < b.QueryName {
				return -1
			} else {
				return anyutil.If(a.QueryName > b.QueryName, 1, 0)
			}
		})
		tables = append(tables, newTables...)
		// Remove the found tables
		for _, t := range newTables {
			unusedTables.Delete(t)
		}
		if unusedTables.Len() == 0 {
			break
		}
		if len(newTables) == 0 {
			// circular references are what is left, so just add everything and return
			tables = append(tables, unusedTables.Values()...)
			break
		}
	}

	return
}

// UniqueManyManyReferences returns all the many-many references, but returning only one per association table.
func (m *Database) UniqueManyManyReferences() []*ManyManyReference {
	refs := make(map[string]*ManyManyReference)
	for _, table := range slices.SortedFunc(maps2.Values(m.Tables), func(table *Table, table2 *Table) int {
		return cmp.Compare(table.QueryName, table2.QueryName)
	}) {
		for _, mm := range table.ManyManyReferences {
			refs[mm.AssnTableName] = mm
		}
	}

	return slices.SortedFunc(maps2.Values(refs), func(reference *ManyManyReference, reference2 *ManyManyReference) int {
		return cmp.Compare(reference.AssnSourceColumnName, reference2.AssnSourceColumnName)
	})
}

package model

import (
	"cmp"
	"github.com/goradd/anyutil"
	"github.com/goradd/maps"
	"github.com/goradd/orm/pkg/db"
	"github.com/goradd/orm/pkg/schema"
	"github.com/kenshaw/snaker"
	"log/slog"
	maps2 "maps"
	"slices"
	"strings"
	"time"
)

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
// schema must be Clean() first.
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
		m.importTable(table, m.WriteTimeout, m.ReadTimeout)
	}

	for _, assn := range schema.AssociationTables {
		m.importAssociation(assn)
	}
}

// Analyzes an association table and creates special virtual columns in the corresponding tables it points to.
// Association tables are used by SQL databases to create many-many relationships. NoSQL databases can define their
// association columns directly and store an array of records on either end of the association.
func (m *Database) importAssociation(schemaAssn *schema.AssociationTable) {
	t1 := m.Table(schemaAssn.Ref1.QualifiedTableName())
	if t1 == nil {
		slog.Error("Missing associated table from association "+schemaAssn.Table,
			slog.String(db.LogTable, schemaAssn.Ref1.Table))
		return
	}
	t2 := m.Table(schemaAssn.Ref2.QualifiedTableName())
	if t2 == nil {
		slog.Error("Missing associated table from association "+schemaAssn.Ref2.Table,
			slog.String(db.LogTable, schemaAssn.Ref2.Table))
		return
	}

	ref1 := makeManyManyRef(
		schemaAssn.Table,
		schemaAssn.Ref1.Column,
		schemaAssn.Ref2.Column,
		t1,
		t2,
		schemaAssn.Ref2.Label,
		schemaAssn.Ref2.LabelPlural,
		schemaAssn.Ref2.Identifier,
		schemaAssn.Ref2.IdentifierPlural,
	)
	ref2 := makeManyManyRef(
		schemaAssn.Table,
		schemaAssn.Ref2.Column,
		schemaAssn.Ref1.Column,
		t2,
		t1,
		schemaAssn.Ref1.Label,
		schemaAssn.Ref1.LabelPlural,
		schemaAssn.Ref1.Schema,
		schemaAssn.Ref1.IdentifierPlural,
	)
	ref1.MM = ref2
	ref2.MM = ref1
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
				if !slices.Contains(tables, ref.ReferencedTable) &&
					!slices.Contains(newTables, ref.ReferencedTable) &&
					t != ref.ReferencedTable {
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
			refs[mm.TableQueryName] = mm
		}
	}

	return slices.SortedFunc(maps2.Values(refs), func(reference *ManyManyReference, reference2 *ManyManyReference) int {
		return cmp.Compare(reference.SourceColumnName, reference2.SourceColumnName)
	})
}

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

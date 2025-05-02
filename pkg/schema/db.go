package schema

import (
	"github.com/goradd/all"
	"github.com/goradd/maps"
	strings2 "github.com/goradd/strings"
	"log/slog"
	"slices"
)

const Version = 1

// Database is a description of the structure of the data in a database that is agnostic of the type
// of the database, including whether its SQL or NoSQL.
type Database struct {
	// Key is the database key corresponding to its key in the global database cluster.
	// Should be unique among the other databases in use.
	Key string `json:"key"`

	// Package is the name of the package and directory that will be created to hold the generated code.
	// Should be all lower case, with no hyphens or underscores. Will be based on Key if omitted.
	Package string `json:"package,omitempty"`

	// ReferenceSuffix is the ending to expect at the end of column names that are referencing other
	// tables (also known as foreign keys in SQL databases).
	// Defaults to "_id". Will be added to references if missing.
	ReferenceSuffix string `json:"reference_suffix,omitempty"`

	// EnumTableSuffix is the text to strip off the end of an enum table name when converting it to a type name.
	// Defaults to "_enum". Will be added to enum table names if missing.
	EnumTableSuffix string `json:"enum_table_suffix,omitempty"`

	// AssnTableSuffix is the suffix for association table names.
	AssnTableSuffix string `json:"assn_table_suffix,omitempty"`

	// Tables are the standard tables in the database.
	Tables []*Table `json:"tables"`

	// EnumTables contains a description of the enumerated types from the enum tables in the database.
	EnumTables []*EnumTable `json:"enum_tables"`

	// AssociationTables form many-to-many relationships between tables in the database.
	AssociationTables []*AssociationTable `json:"association_tables"`
}

// FillDefaults will fill all the undeclared values in the database structure with default values.
func (db *Database) FillDefaults() {
	if db.Package == "" {
		db.Package = db.Key
	}
	// remove invalid characters
	s := SanitizePackageName(db.Package)
	if s != db.Package {
		slog.Warn("Package name was modified",
			slog.String("original name", db.Package),
			slog.String("new name", s),
		)
		db.Package = s
	}

	if db.ReferenceSuffix == "" {
		db.ReferenceSuffix = "_id"
	}
	if db.EnumTableSuffix == "" {
		db.EnumTableSuffix = "_enum"
	}
	if db.AssnTableSuffix == "" {
		db.AssnTableSuffix = "_assn"
	}

	for _, t := range db.Tables {
		t.FillDefaults(db)
	}

	for _, t := range db.EnumTables {
		t.FillDefaults(db.EnumTableSuffix)
	}

	for _, t := range db.AssociationTables {
		t.FillDefaults(db.ReferenceSuffix)
	}
	db.Clean()
}

// FindTable finds the table by name. Returns nil if not found.
// name should be schema.table if the table has a schema specified.
func (db *Database) FindTable(name string) *Table {
	for _, t := range db.Tables {
		if t.QualifiedName() == name {
			return t
		}
	}
	return nil
}

// FindEnumTable finds the enum table by name. Returns nil if not found.
// name should be schema.table if the table has a schema specified.
func (db *Database) FindEnumTable(name string) *EnumTable {
	for _, t := range db.EnumTables {
		if t.QualifiedName() == name {
			return t
		}
	}
	return nil
}

// Clean modifies the structure to prepare it for creating a schema in a database.
// A cleaned structure should be saved so that it can be synchronized with the database as it changes.
func (db *Database) Clean() {
	db.sort()
}

// FillKeys assigns keys to all parts of the schema as an aid in synchronization when the schema changes.
func (db *Database) FillKeys() {
	for _, t := range db.Tables {
		if t.Key == "" {
			t.Key = strings2.RandomString(strings2.AlphaNumeric, 10)
		}
		for _, c := range t.Columns {
			if c.Key == "" {
				c.Key = strings2.RandomString(strings2.AlphaNumeric, 10)
			}
		}
	}

	for _, t := range db.EnumTables {
		if t.Key == "" {
			t.Key = strings2.RandomString(strings2.AlphaNumeric, 10)
		}
	}

	for _, t := range db.AssociationTables {
		if t.Key == "" {
			t.Key = strings2.RandomString(strings2.AlphaNumeric, 10)
		}
	}
}

// sort will sort the Tables, EnumTables and AssociationTables into a predictable order that also
// orders the tables so that earlier tables do not reference later tables.
func (db *Database) sort() {
	var unusedTables maps.SliceSet[*Table]

	unusedTables.SetSortFunc(func(a, b *Table) bool {
		return a.Name < b.Name
	})
	unusedTables.Add(db.Tables...)
	db.Tables = nil
	for { // repeat until unusedTables is empty
		var newTables []*Table
	nexttable:
		for t := range unusedTables.All() {
			for _, col := range t.Columns {
				if col.Type == ColTypeReference {
					// skip this table if it has references to a table we have not yet seen
					if !slices.ContainsFunc(db.Tables, func(t2 *Table) bool { return t2.Name == col.Reference.Table }) &&
						!slices.ContainsFunc(newTables, func(t2 *Table) bool { return t2.Name == col.Reference.Table }) &&
						t.Name != col.Reference.Table {
						continue nexttable
					}
				}
			}
			// This has no forward references we care about
			newTables = append(newTables, t)
		}
		slices.SortFunc(newTables, func(a, b *Table) int {
			if a.Name < b.Name {
				return -1
			} else {
				return all.If(a.Name > b.Name, 1, 0)
			}
		})
		db.Tables = append(db.Tables, newTables...)
		// Remove the found tables
		for _, t := range newTables {
			unusedTables.Delete(t)
		}
		if unusedTables.Len() == 0 {
			break
		}
		if len(newTables) == 0 {
			// circular references are what is left, so just add everything and return
			db.Tables = append(db.Tables, unusedTables.Values()...)
			break
		}
	}

	slices.SortFunc(db.EnumTables, func(a, b *EnumTable) int {
		if a.Name < b.Name {
			return -1
		} else {
			return all.If(a.Name > b.Name, 1, 0)
		}
	})

	slices.SortFunc(db.AssociationTables, func(a, b *AssociationTable) int {
		if a.Name < b.Name {
			return -1
		} else {
			return all.If(a.Name > b.Name, 1, 0)
		}
	})

	return
}

package schema

import (
	"github.com/goradd/anyutil"
	"github.com/goradd/maps"
	"log/slog"
	"slices"
)

const Version = 1

// Database is a description of the structure of the data in a database that is agnostic of the type
// of the database, including whether its SQL or NoSQL.
//
// The purpose is to create a structure that is as easy as possible to be specified by humans,
// with many of the values being optional as they can be inferred from other values.
//
// See model.Database for the structure that is presented to the code generator and that is based on this structure.
type Database struct {
	// Key is the database key corresponding to its key in the global database cluster.
	// Should be unique among the other databases in use.
	Key string `json:"key"`

	// Package is the name of the root package for top level code sent to the output directory.
	// Should be all lower case, with no hyphens or underscores.
	// Will be the name of the output directory if omitted.
	Package string `json:"package,omitempty"`

	// ImportPath is the import path that refers to the output directory.
	// This value will be used to refer to files across packages.
	// If omitted, it will be calculated based on the output directory by looking for the go.mod
	// file that applies to the output directory.
	ImportPath string `json:"import_path,omitempty"`

	// WriteTimeout is used to wrap write transactions with a timeout on their contexts.
	// Leaving it as zero will turn off timeout wrapping, allowing you to wrap individual calls as you
	// see fit. This only applies to code generated transactions. See also table.WriteTimeout to set
	// a timeout value on an individual table.
	// Use a duration format understood by time.ParseDuration.
	WriteTimeout string `json:"write_timeout,omitempty"`

	// ReadTimeout is used to wrap read transactions with a timeout on their contexts.
	// Leaving it as zero will turn off timeout wrapping, allowing you to wrap individual calls as you
	// see fit. This only applies to code generated transactions. See also table.ReadTimeout to set
	// a timeout value on an individual table.
	// Use a duration format understood by time.ParseDuration.
	ReadTimeout string `json:"read_timeout,omitempty"`

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

// FillDefaults will fill all the undeclared values in the database structure with default values
// in preparation for building the model.Database structure.
func (db *Database) FillDefaults() {
	// remove invalid characters
	s := SanitizePackageName(db.Package)
	if s != db.Package {
		slog.Warn("Package name was modified",
			slog.String("original name", db.Package),
			slog.String("new name", s),
		)
		db.Package = s
	}

	if db.EnumTableSuffix == "" {
		db.EnumTableSuffix = "_enum"
	}
	if db.AssnTableSuffix == "" {
		db.AssnTableSuffix = "_assn"
	}

	for _, t := range db.Tables {
		t.fillDefaults(db)
	}

	for _, t := range db.EnumTables {
		t.fillDefaults(db.EnumTableSuffix)
	}

	for _, t := range db.AssociationTables {
		t.fillDefaults(db)
	}
}

// infer fills in certain key inferred values. The goal is to infer the minimal set of values
// to create a database structure. There is some coordination here with the SQL databases that can
// export their schemas.
// It also does some validity checks.
func (db *Database) infer() error {
	for _, t := range db.Tables {
		if err := t.Clean(db); err != nil {
			return err
		}
	}

	for _, t := range db.AssociationTables {
		if err := t.infer(db); err != nil {
			return err
		}
	}

	for _, t := range db.EnumTables {
		if err := t.infer(db); err != nil {
			return err
		}
	}

	return nil
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
		if t.QualifiedTableName() == name {
			return t
		}
	}
	return nil
}

// Clean modifies the structure to prepare it for creating a schema in a database.
func (db *Database) Clean() error {
	db.Sort() // required before infer so references work
	if err := db.infer(); err != nil {
		return err
	}
	return nil
}

// Sort will Sort the Tables, EnumTables and AssociationTables into a predictable order that also
// orders the tables so that earlier tables do not reference later tables.
func (db *Database) Sort() {
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
			for _, ref := range t.References {
				// skip this table if it has references to a table we have not yet seen
				if !slices.ContainsFunc(db.Tables, func(t2 *Table) bool { return t2.Name == ref.Table }) &&
					!slices.ContainsFunc(newTables, func(t2 *Table) bool { return t2.Name == ref.Table }) &&
					t.Name != ref.Table {
					continue nexttable
				}
			}
			// This has no forward references we care about
			newTables = append(newTables, t)
		}
		slices.SortFunc(newTables, func(a, b *Table) int {
			if a.Name < b.Name {
				return -1
			} else {
				return anyutil.If(a.Name > b.Name, 1, 0)
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
			// circular references are what is left
			panic("database has circular primary key references")
		}
	}

	slices.SortFunc(db.EnumTables, func(a, b *EnumTable) int {
		if a.QualifiedTableName() < b.QualifiedTableName() {
			return -1
		} else {
			return anyutil.If(a.QualifiedTableName() > b.QualifiedTableName(), 1, 0)
		}
	})

	slices.SortFunc(db.AssociationTables, func(a, b *AssociationTable) int {
		if a.QualifiedTableName() < b.QualifiedTableName() {
			return -1
		} else {
			return anyutil.If(a.QualifiedTableName() > b.QualifiedTableName(), 1, 0)
		}
	})

	return
}

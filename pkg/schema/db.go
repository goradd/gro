package schema

import (
	"strings"
	"unicode"
)

const Version = 1

// Database is the description of a single database in a database type agnostic way.
// A database can have foreign keys making connections between tables within the database.
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

	// AssociationTables form many-to-many relationships between tables in the database
	AssociationTables []*AssociationTable `json:"association_tables"`
}

// FillDefaults will fill all the undeclared values in the database structure with default values.
func (db *Database) FillDefaults() {
	if db.Package == "" {
		db.Package = strings.ToLower(db.Key)
		// remove anything not a letter or number
		db.Package = strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return r
			}
			return -1
		}, db.Package)
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
		t.FillDefaults(db.ReferenceSuffix, db.EnumTableSuffix)
	}

	for _, t := range db.EnumTables {
		t.FillDefaults(db.EnumTableSuffix)
	}
}

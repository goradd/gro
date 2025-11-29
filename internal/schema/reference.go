package schema

import (
	"fmt"
	"log/slog"
	"strings"

	strings2 "github.com/goradd/strings"
	"github.com/kenshaw/snaker"
)

// Reference represents a forward link to an object described by a table.
// In SQL databases these are known as foreign-keys.
// The link will be a one-to-one link if IsUnique is true, or a one-to-many link.
// Reference will create a column in this table that holds a duplicate of the primary key
// from the referenced table.
// Reference automatically creates a reverse link in the referenced table. That reverse link may
// create a column in the other table, or may be a kind of virtual column that only gets populated during
// a query using an index.
//
// References to tables with composite keys are not currently supported.
type Reference struct {
	// Table is the name of the table being referenced.
	// If Table is the same as the column's table, it creates a parent-child relationship.
	// Should match a Table.Name value in another table.
	Table string `json:"table"`

	// Schema is the schema that Table belongs to. If empty, will use the default schema.
	Schema string `json:"schema,omitempty"`

	// Column is the name of the local column created in this table to hold a duplicate of the private key in Table,
	// and is also known as a foreign key.
	// It will default to a name based on Table and the name of the primary key column in Table.
	Column string `json:"column,omitempty"`

	// ColumnIdentifier is the name that Go will use to identify the column
	ColumnIdentifier string `json:"column_identifier,omitempty"`

	// ColumnLabel is the human-readable name of the column.
	ColumnLabel string `json:"column_label,omitempty"`

	/* Future expansion for composite keys:
	// Maps a column name in the composite primary key in Table to a local column.
	// If empty, will infer the values from the names of the columns in Table.
	// User should fill out Column for single key tables, or Columns for composite-key tables.
	Columns map[string]ColumnDescription // (name, identifier, label)
	*/

	// ObjectIdentifier is the Go name used for the referenced object.
	// If not specified, will be based on Column.
	ObjectIdentifier string `json:"object_identifier,omitempty"`

	// Label is the human-readable name for the referenced object.
	// If not specified, will be based on Identifier.
	ObjectLabel string `json:"object_label,omitempty"`

	// IndexLevel specifies the index to be created on the foreign key column.
	// If a primary key or unique key, will create a one-to-one relationship.
	// If left empty, will default to IndexLevelIndexed.
	IndexLevel IndexLevel `json:"index_level,omitempty"`

	// IsNullable indicates that the reference is not required and can be represented as a null value in the database.
	// If not nullable, then the reference is required to be present and valid when the record is saved.
	// This would also mean that if the referenced object is deleted, or changes its reverse
	// relationship, then this object will be deleted.
	// The foreign key column will be nullable if this is nullable.
	IsNullable bool `json:"nullable,omitempty"`

	// The singular Go identifier that will be used for the reverse relationship objects.
	// If not specified, will be based on Table.Name.
	// Should be CamelCase with no spaces.
	// For example, "ManagedProject".
	ReverseIdentifier string `json:"reverse_identifier,omitempty"`

	// The plural Go identifier that will be used for the reverse relationship objects.
	// If not specified, the ReverseIdentifier will be pluralized.
	ReverseIdentifierPlural string `json:"reverse_identifier_plural,omitempty"`

	// The singular description of the Table objects as referred to by the referenced table.
	// If not specified, will be based on ReverseIdentifier.
	// For example, "Managed Project", or "Project I Manage".
	ReverseLabel string `json:"reverse_label,omitempty"`

	// The plural description of Table objects as referred to by the referenced table.
	// If not specified, the ReverseLabel will be pluralized.
	ReverseLabelPlural string `json:"reverse_label_plural,omitempty"`
}

func (r *Reference) infer(db *Database, table *Table) error {

	// Double check some of the values
	if r.Table == "" {
		slog.Error("Table value in Reference was not specified",
			slog.String("table", table.Name))
		return fmt.Errorf("table value in Reference was not specified in table %s", table.Name)
	}
	t := db.FindTable(r.Table)
	if t == nil {
		slog.Error("Referenceed Table was not found in the database",
			slog.String("referring table", table.Name),
			slog.String("referenced table", r.Table))
		return fmt.Errorf("table %s was not found", r.Table)
	}
	// Find the primary key column in the other table.
	// If the other table does not have a set single primary key column, then a name cannot be inferred
	cols := t.PrimaryKeyColumns()
	if len(cols) != 1 {
		// We do not currently support references to tables with composite primary keys or with no key.
		slog.Error("Referenced table does not have a single, non-referenced primary key, so column name cannot be inferred.",
			slog.String("referring table", table.Name),
			slog.String("referenced table", r.Table),
		)
		return fmt.Errorf("referenced table %s does not have a single primary key, so column name cannot be inferred. ", r.Table)
	}
	pk := t.FindColumn(cols[0])
	if pk == nil {
		return fmt.Errorf("primary key column %s not found", cols[0])
	}

	if r.Column == "" {
		r.Column = r.Table + "_" + pk.Name
	}

	if r.IndexLevel == IndexLevelNone {
		r.IndexLevel = IndexLevelIndexed
	}
	return nil
}

// fillDefaults fills optional default values.
// It expects infer to have already been called.
func (r *Reference) fillDefaults(db *Database, table *Table) {
	if r.ColumnIdentifier == "" {
		r.ColumnIdentifier = snaker.SnakeToCamelIdentifier(r.Column)
	}
	if r.ColumnLabel == "" {
		r.ColumnLabel = strings2.Title(r.ColumnIdentifier)
	}
	if r.ObjectIdentifier == "" {
		t := db.FindTable(r.Table)
		pks := t.PrimaryKeyColumns()
		if len(pks) == 1 {
			objName := strings.TrimSuffix(r.Column, "_"+pks[0])
			r.ObjectIdentifier = snaker.SnakeToCamelIdentifier(objName)
		} else {
			r.ObjectIdentifier = snaker.SnakeToCamelIdentifier(r.Table)
		}
	}
	if r.ObjectLabel == "" {
		r.ObjectLabel = strings2.Title(r.ObjectIdentifier)
	}

	if r.ReverseIdentifier == "" {
		if r.ObjectIdentifier == snaker.SnakeToCamelIdentifier(r.Table) {
			r.ReverseIdentifier = table.Identifier
		} else {
			r.ReverseIdentifier = r.ObjectIdentifier + table.Identifier
		}
	}
	if r.ReverseIdentifierPlural == "" {
		r.ReverseIdentifierPlural = strings2.Plural(r.ReverseIdentifier)
	}

	if r.ReverseLabel == "" {
		r.ReverseLabel = strings2.Title(r.ObjectIdentifier)
	}
	if r.ReverseLabelPlural == "" {
		r.ReverseLabelPlural = strings2.Plural(r.ReverseLabel)
	}
}

// ReferenceColumns generates a foreign key column from the info in the reference
// and returns the primary key column it refers to.
func (r *Reference) ReferenceColumns(db *Database, table *Table) (*Column, *Column) {
	t := db.FindTable(r.Table)
	// Find the primary key column in the other table.
	// If the other table does not have a set single primary key column, then a name cannot be inferred
	cols := t.PrimaryKeyColumns()
	if len(cols) != 1 {
		panic(fmt.Errorf("referenced table %s does not have a single primary key, so column name cannot be inferred. ", r.Table))
	}
	pk := t.FindColumn(cols[0])
	if pk == nil {
		panic(fmt.Errorf("primary key column %s not found", cols[0]))
	}
	// generate a column for the reference
	fk := *pk
	fk.Name = r.Column
	fk.IndexLevel = IndexLevelNone // indexing is handled by the reference
	fk.Identifier = r.ColumnIdentifier
	fk.Label = r.ColumnLabel
	fk.Comment = ""
	fk.EnumTable = ""
	fk.IsNullable = r.IsNullable

	return &fk, pk
}

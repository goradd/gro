package model

import (
	"github.com/goradd/anyutil"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"log/slog"
)

// Reference describes a forward relationship.
// Cross database references are not supported.
// References will cause a copy of the primary key of Table to be placed in Column,
// and will generate code that refers to an object in Table as Identifier, and reverse
// code in Table that will refer to objects in this table as ReverseIdentifier.
type Reference struct {
	// Table is the referenced table.
	Table *Table
	// Column is the local column that is referring to the referenced table
	Column *Column
	// The go name of the forward referenced object
	Identifier string
	// The local name used to refer to the referenced object
	DecapIdentifier string
	// The label of the object referred to.
	Label string
	// ReverseLabel is the human-readable label of the object of the reverse relationship.
	ReverseLabel string
	// ReverseLabelPlural is the plural of ReverseLabel.
	ReverseLabelPlural string
	// ReverseIdentifier is the name we should use to refer to the related object.
	ReverseIdentifier string
	// ReverseIdentifierPlural is the name we should use to refer to the plural of the related object.
	ReverseIdentifierPlural string
	// IsUnique indicates that this is a one-to-one relationship
	IsUnique bool
}

// VariableIdentifier returns the name of the local variable that will
// hold the object loaded in the reference.
func (r *Reference) VariableIdentifier() string {
	return "obj" + r.Identifier
}

// GoType returns the name of the Go struct type in a forward reference.
func (r *Reference) GoType() string {
	return r.Column.Table.Identifier
}

// ReverseVariableIdentifier returns the name of the local variable that will
// hold the object(s) loaded in the reverse reference.
func (r *Reference) ReverseVariableIdentifier() string {
	if r.IsUnique {
		return "rev" + r.ReverseIdentifier
	} else {
		return "rev" + r.ReverseIdentifierPlural
	}
}

// JsonKey returns the key that will be used for the referenced object in JSON.
func (r *Reference) JsonKey() string {
	return r.DecapIdentifier
}

// ReverseJsonKey returns the key that will be used for the reverse referenced object in JSON.
func (r *Reference) ReverseJsonKey() string {
	if r.IsUnique {
		return LowerCaseIdentifier(r.ReverseIdentifier)
	} else {
		return LowerCaseIdentifier(r.ReverseIdentifierPlural)
	}
}

// importReference creates a reference from a schemaRef.
func (m *Database) importReference(schemaRef *schema.Reference) *Reference {
	refTable := m.Table(schemaRef.Table)
	if refTable == nil {
		slog.Error("Referenced table does not exist",
			slog.String(schemaRef.Table, schemaRef.Table))
		return nil
	}
	pk := refTable.PrimaryKeyColumn()
	if pk == nil {
		slog.Warn("Referenced table does not have a single primary key column.",
			slog.String(schemaRef.Table, schemaRef.Table))
		return nil
	}

	col := &Column{
		Table:      refTable,
		QueryName:  schemaRef.Column,
		Identifier: schemaRef.ColumnIdentifier,
		Label:      schemaRef.ColumnLabel,
		SchemaType: anyutil.If(pk.SchemaType == schema.ColTypeAutoPrimaryKey,
			schema.ColTypeString,
			pk.SchemaType),
		SchemaSubType:   pk.SchemaSubType,
		ReceiverType:    pk.ReceiverType,
		Size:            pk.Size,
		DefaultValue:    pk.DefaultValue,
		IsNullable:      schemaRef.IsNullable,
		goType:          pk.ReceiverType.GoType(),
		decapIdentifier: strings2.Decap(schemaRef.ColumnIdentifier),
	}

	ref := &Reference{
		Table:                   refTable,
		Identifier:              schemaRef.ObjectIdentifier,
		Label:                   schemaRef.ObjectLabel,
		ReverseLabel:            schemaRef.ReverseLabel,
		ReverseLabelPlural:      schemaRef.ReverseLabelPlural,
		ReverseIdentifier:       schemaRef.ReverseIdentifier,
		ReverseIdentifierPlural: schemaRef.ReverseIdentifierPlural,
		DecapIdentifier:         strings2.Decap(schemaRef.ObjectIdentifier),
		Column:                  col,
		IsUnique:                schemaRef.IndexLevel == schema.IndexLevelPrimaryKey || schemaRef.IndexLevel == schema.IndexLevelUnique,
	}

	refTable.ReverseReferences = append(refTable.ReverseReferences, ref)
	col.Reference = ref
	return ref
}

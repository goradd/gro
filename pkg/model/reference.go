package model

import (
	"github.com/goradd/anyutil"
	"github.com/goradd/orm/pkg/schema"
	strings2 "github.com/goradd/strings"
	"log/slog"
)

// Reference describes a forward relationship to a related Table.
// Relationships are through a foreign key column in this table, that contains a copy of the
// primary key from the related table.
// Cross database references are not supported.
// References will generate code that refers to an object through the accessor function Identifier,
// and reverse code in the referenced Table that will refer to objects in this table as ReverseIdentifier.
type Reference struct {
	// Table is a pointer back to the table this reference is part of.
	Table *Table
	// ReferencedTable is the table being pointed to by ForeignKey.
	ReferencedTable *Table
	// ForeignKey is the local column that is referring to the referenced table
	ForeignKey *Column
	// Identifier is the go identifier used as an accessor to the forward referenced object.
	// Example: ProjectManager
	Identifier string
	// Field is the name used in the struct representing a pointer to the referenced object.
	// Also used as a function parameter name.
	// Example: projectManager
	Field string
	// The human-readable label of the object referred to.
	// Example: Project Manager
	Label string
	// ReverseIdentifier is the name we should use to refer to the related object.
	// Example: Project
	ReverseIdentifier string
	// ReverseIdentifierPlural is the name we should use to refer to the plural of the related object.
	// Example: Projects
	ReverseIdentifierPlural string
	// IsUnique indicates that this is a one-to-one relationship
	// ReverseLabel is the human-readable label of the object of the reverse relationship.
	// Example: Project
	ReverseLabel string
	// ReverseLabelPlural is the plural of ReverseLabel.
	// Example: Projects
	ReverseLabelPlural string
	// ReverseField is the name used in the struct for the reverse object(s)
	ReverseField string
	// IsUnique is true if this is a one-to-one relationship
	IsUnique bool
	// IsNullable is true if the foreign keys are nullable. This also indicates that the
	// reference is not required when the table is saved.
	IsNullable bool
}

// JsonKey returns the key that will be used for the referenced object in JSON.
func (r *Reference) JsonKey() string {
	return r.Field
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
func (m *Database) importReference(schemaTable *schema.Table, schemaRef *schema.Reference) *Reference {
	table := m.Table(schemaTable.Name)
	if table == nil {
		slog.Error("Table does not exist",
			slog.String(schemaRef.Table, schemaTable.Name))
		return nil
	}

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
		Table:      table,
		QueryName:  schemaRef.Column,
		Identifier: schemaRef.ColumnIdentifier,
		Label:      schemaRef.ColumnLabel,
		SchemaType: anyutil.If(pk.SchemaType == schema.ColTypeAutoPrimaryKey,
			schema.ColTypeString,
			pk.SchemaType),
		SchemaSubType: pk.SchemaSubType,
		ReceiverType:  pk.ReceiverType,
		Size:          pk.Size,
		DefaultValue:  pk.DefaultValue,
		IsNullable:    schemaRef.IsNullable,
		Type:          pk.ReceiverType.GoType(),
		Field:         strings2.Decap(schemaRef.ColumnIdentifier),
		FieldPlural:   strings2.Plural(strings2.Decap(schemaRef.ColumnIdentifier)),
	}

	ref := &Reference{
		ForeignKey:              col,
		Table:                   table,
		ReferencedTable:         refTable,
		Identifier:              schemaRef.ObjectIdentifier,
		Field:                   strings2.Decap(schemaRef.ObjectIdentifier),
		Label:                   schemaRef.ObjectLabel,
		ReverseLabel:            schemaRef.ReverseLabel,
		ReverseLabelPlural:      schemaRef.ReverseLabelPlural,
		ReverseIdentifier:       schemaRef.ReverseIdentifier,
		ReverseIdentifierPlural: schemaRef.ReverseIdentifierPlural,
		ReverseField:            strings2.Decap(schemaRef.ReverseIdentifier),
		IsUnique:                schemaRef.IndexLevel == schema.IndexLevelPrimaryKey || schemaRef.IndexLevel == schema.IndexLevelUnique,
		IsNullable:              schemaRef.IsNullable,
	}

	refTable.ReverseReferences = append(refTable.ReverseReferences, ref)
	col.Reference = ref
	return ref
}

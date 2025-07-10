package model

import (
	"github.com/goradd/orm/pkg/query"
	"github.com/goradd/strings"
)

// The ManyManyReference structure is used by the templates during the codegen process to describe
// a many-to-many relationship.
// Underlying the structure is an association table that has two foreign keys pointing
// to the records that are linked.
// For each relationship, two ManyManyReference structures are created.
type ManyManyReference struct {
	// TableQueryName is the database table that links the two associated tables together.
	TableQueryName string
	// ForeignKeyName is the database column in the association table that points at the referenced table's primary key.
	ForeignKeyName string
	// ForeignKeyReceiverType is the query.ReceiverType of the destination column.
	ForeignKeyReceiverType query.ReceiverType
	// ReferencedTable is the table being linked.
	ReferencedTable *Table

	// Label is the human-readable label of the objects pointed to.
	Label string
	// LabelPlural is the plural of Label
	LabelPlural string
	// Identifier is the name used to refer to an object on the other end of the reference.
	// It is not the same as the object type. For example TeamMember would refer to a Person type.
	// This is derived from the ForeignKeyName but can be overridden.
	Identifier string
	// IdentifierPlural is the name used to refer to the group of objects on the other end of the reference.
	// For example, TeamMembers. This is derived from the ForeignKeyName but can be overridden by
	// a comment in the table.
	IdentifierPlural string
	// Field is the go identifier that will be used in the Table struct and parameters.
	// Since this always points to many objects, it will be a plural name.
	Field string

	// MM is the many-many reference on the other end of the relationship that points back to this one.
	MM *ManyManyReference
}

// JsonKey returns the key used when referring to the associated objects in JSON.
func (m *ManyManyReference) JsonKey() string {
	return LowerCaseIdentifier(m.IdentifierPlural)
}

// Type returns the name of the object type the association links to.
func (m *ManyManyReference) Type() string {
	return m.ReferencedTable.Identifier
}

// TypePlural returns the plural name of the object type the association links to.
func (m *ManyManyReference) TypePlural() string {
	return m.ReferencedTable.IdentifierPlural
}

// PrimaryKeyColumnName returns the database name of the primary key of the object the association links to.
func (m *ManyManyReference) PrimaryKeyColumnName() string {
	return m.ReferencedTable.PrimaryKeyColumn().QueryName
}

// PrimaryKeyType returns the type of the referenced table's primary key.
func (m *ManyManyReference) PrimaryKeyType() string {
	return m.ReferencedTable.PrimaryKeyType()
}

// TableIdentifier identifies the association table.
func (m *ManyManyReference) TableIdentifier() string {
	return UpperCaseIdentifier(m.TableQueryName)
}

// PkField returns the identifier used for the variable listing the primary keys of the associated objects.
func (m *ManyManyReference) PkField() string {
	return m.Field + "Pks"
}

func (m *ManyManyReference) SourceColumnName() string {
	return m.MM.ForeignKeyName
}

func (m *ManyManyReference) SourceColumnReceiverType() query.ReceiverType {
	return m.MM.ForeignKeyReceiverType
}

func (m *ManyManyReference) SourcePrimaryKeyName() string {
	return m.MM.ReferencedTable.PrimaryKeyColumn().QueryName
}

func makeManyManyRef(assnTable, fk string, refTable *Table, label, labels, id, ids string) *ManyManyReference {
	pk := refTable.PrimaryKeyColumn()

	ref := ManyManyReference{
		TableQueryName:         assnTable,
		ForeignKeyName:         fk,
		ForeignKeyReceiverType: pk.ReceiverType,
		ReferencedTable:        refTable,
		Label:                  label,
		LabelPlural:            labels,
		Identifier:             id,
		IdentifierPlural:       ids,
		Field:                  strings.Decap(ids),
	}
	return &ref
}

package model

import "github.com/goradd/orm/pkg/query"

// The ManyManyReference structure is used by the templates during the codegen process to describe a many-to-many relationship.
// Underlying the structure is an association table that has two foreign keys pointing
// to the records that are linked.
// For each relationship, two ManyManyReference structures are created.
type ManyManyReference struct {
	// AssnTableName is the database table that links the two associated tables together.
	AssnTableName string
	// AssnSourceColumnName is the database column in the association table that points at the source table's primary key.
	AssnSourceColumnName string
	// AssnSourceColumnType is the type of the column in the association table.
	AssnSourceColumnType query.ReceiverType
	// AssnDestColumnName is the database column in the association table that points at the destination table's primary key.
	AssnDestColumnName string
	// AssnDestColumnType is the type of the column in the association table.
	AssnDestColumnType query.ReceiverType
	// DestinationTable is the table being linked (the table that we are joining to)
	DestinationTable *Table

	// Label is the human-readable label of the objects pointed to.
	Label string
	// LabelPlural is the plural of Label
	LabelPlural string
	// Identifier is the name used to refer to an object on the other end of the reference.
	// It is not the same as the object type. For example TeamMember would refer to a Person type.
	// This is derived from the AssnDestColumnName but can be overridden.
	Identifier string
	// IdentifierPlural is the name used to refer to the group of objects on the other end of the reference.
	// For example, TeamMembers. This is derived from the AssnDestColumnName but can be overridden by
	// a comment in the table.
	IdentifierPlural string

	// MM is the many-many reference on the other end of the relationship that points back to this one.
	MM *ManyManyReference
}

// TableName returns the name of the association table. This is mainly used to import and export
// the table.
func (m *ManyManyReference) TableName() string {
	return UpperCaseIdentifier(m.AssnTableName)
}

// JsonKey returns the key used when referring to the associated objects in JSON.
func (m *ManyManyReference) JsonKey() string {
	return LowerCaseIdentifier(m.IdentifierPlural)
}

// ObjectType returns the name of the object type the association links to.
func (m *ManyManyReference) ObjectType() string {
	return m.DestinationTable.Identifier
}

// ObjectTypePlural returns the plural name of the object type the association links to.
func (m *ManyManyReference) ObjectTypePlural() string {
	return m.DestinationTable.IdentifierPlural
}

// PrimaryKeyType returns the Go type of the primary key of the object the association links to.
func (m *ManyManyReference) PrimaryKeyType() string {
	return m.DestinationTable.PrimaryKeyGoType()
}

// PrimaryKey returns the database field name of the primary key of the object the association links to.
func (m *ManyManyReference) PrimaryKey() string {
	return m.DestinationTable.PrimaryKeyColumn().QueryName
}

// QueryName returns the database table name of the destination table.
func (m *ManyManyReference) QueryName() string {
	return m.DestinationTable.QueryName
}

// VariableIdentifier is the local variable name used to identify queried objects attached to the local object
// through the many-many relationship.
func (m *ManyManyReference) VariableIdentifier() string {
	return "mm" + m.IdentifierPlural
}

func (m *ManyManyReference) PkIdentifier() string {
	return "mm" + m.IdentifierPlural + "Pks"
}

func makeManyManyRef(
	assnTable string,
	column1, column2 string,
	t1, t2 *Table,
	label, labels, id, ids string,
) *ManyManyReference {
	pk1 := t1.PrimaryKeyColumn()
	type1 := pk1.ReceiverType

	pk2 := t2.PrimaryKeyColumn()
	type2 := pk2.ReceiverType

	ref := ManyManyReference{
		AssnTableName:        assnTable,
		AssnSourceColumnName: column1,
		AssnSourceColumnType: type1,
		AssnDestColumnName:   column2,
		AssnDestColumnType:   type2,
		DestinationTable:     t2,
		Label:                label,
		LabelPlural:          labels,
		Identifier:           id,
		IdentifierPlural:     ids,
	}
	t1.ManyManyReferences = append(t1.ManyManyReferences, &ref)
	return &ref
}

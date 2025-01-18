package model

import "github.com/goradd/orm/pkg/query"

// The ManyManyReference structure is used by the templates during the codegen process to describe a many-to-many relationship.
// Underlying the structure is an association table that has two values that are foreign keys pointing
// to the records that are linked. The names of these fields will determine the names of the corresponding accessors
// in each of the model objects. This allows multiple of these many-many relationships to exist
// on the same tables but for different purposes. These fields MUST point to primary keys.
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

	// Title is the human-readable title of the objects pointed to.
	Title string
	// TitlePlural is the plural of Title
	TitlePlural string
	// Identifier is the name used to refer to an object on the other end of the reference.
	// It is not the same as the object type. For example TeamMember would refer to a Person type.
	// This is derived from the AssnDestColumnName but can be overridden.
	Identifier string
	// IdentifierPlural is the name used to refer to the group of objects on the other end of the reference.
	// For example, TeamMembers. This is derived from the AssnDestColumnName but can be overridden by
	// a comment in the table.
	IdentifierPlural string

	// MM is the many-many reference on the other end of the relationship that points back to this one.
	// This will be nil if the other side is an enum table.
	MM *ManyManyReference
}

// TableIdentifier returns a name to use to describe the relationship in code.
func (m *ManyManyReference) TableIdentifier() string {
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
	assnTable, column1, column2 string,
	t1, t2 *Table,
	title, titles, id, ids string,
) *ManyManyReference {
	type1 := t1.PrimaryKeyColumn().ReceiverType
	type2 := t2.PrimaryKeyColumn().ReceiverType
	ref := ManyManyReference{
		AssnTableName:        assnTable,
		AssnSourceColumnName: column1,
		AssnSourceColumnType: type1,
		AssnDestColumnName:   column2,
		AssnDestColumnType:   type2,
		DestinationTable:     t2,
		Title:                title,
		TitlePlural:          titles,
		Identifier:           id,
		IdentifierPlural:     ids,
	}
	t1.ManyManyReferences = append(t1.ManyManyReferences, &ref)
	return &ref
}

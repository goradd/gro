package model

// Reference describes a forward relationship.
// Cross database references are not supported.
// References will cause a copy of the primary key of Table to be placed in Column,
// and will generate code that refers to an object in Table as Identifier, and reverse
// code in Table that will refer to objects in this table as ReverseIdentifier.
type Reference struct {
	// Table is the referenced table.
	Table *Table
	// Column is the column that is referring to the referenced table
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
}

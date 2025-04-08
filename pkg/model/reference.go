package model

// Reference is additional information to describe what a forward reference points to.
// Cross database references are not supported. References are always to the primary
// key of Table.
type Reference struct {
	// Table is the table on the other end of the foreign key.
	Table *Table
	// If this is a reference to an enum table, EnumTable will point to that enum table
	EnumTable *Enum
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

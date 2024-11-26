package model

// Reference is additional information to describe what a forward reference points to.
// Cross database references are not supported.
type Reference struct {
	// Table is the table on the other end of the foreign key.
	Table *Table
	// Column is the database column in the linked table that matches this column name.
	// Often that is the primary key of the other table.
	Column *Column
	// If this is a reference to an enum table, EnumTable will point to that enum table
	EnumTable *EnumTable
	// The go object name of the forward reference.
	Identifier string
	// The title of the object referred to.
	Title string
	// ReverseTitle is the human-readable title of the object of the reverse relationship.
	ReverseTitle string
	// ReverseTitlePlural is the plural of ReverseTitle.
	ReverseTitlePlural string
	// ReverseIdentifier is the name we should use to refer to the related object.
	ReverseIdentifier string
	// ReverseIdentifierPlural is the name we should use to refer to the plural of the related object.
	ReverseIdentifierPlural string
}

// ReverseVariableIdentifier returns the variable name that generated code will use to refer to objects of this type.
func (r *Reference) ReverseVariableIdentifier() string {
	return "obj" + r.ReverseIdentifier
}

package schema

// Reference is the additional information needed for reference type and enum columns.
// For reference columns, if the IndexLevel of the containing column is Unique, it creates a one-to-one relationship.
// Otherwise, it is a one-to-many relationship.
type Reference struct {
	// If this column is a reference to an object in another table, this is the name of that other table.
	// If using schemas, the format should be "SchemaName.TableName".
	// If Table is the same as the column's table, it creates a parent-child relationship.
	// Should match a Table.Name value in another table.
	// Enum values should point to an enum table.
	Table string `json:"table"`

	// For future expansion. If this is a reference to a table with a composite key, this will
	// specify which specific column is being mirrored. For now, this is unused.
	Column string `json:"column,omitempty"`

	// Identifier is the Go name used for the referenced object.
	// If not specified, will be based on Table.
	Identifier string `json:"identifier,omitempty"`

	// Label is the human-readable name for the referenced object.
	// If not specified, will be based on Identifier.
	Label string `json:"label,omitempty"`

	// The singular Go identifier that will be used for the reverse relationships.
	// If not specified, will be based on Table.Name.
	// Should be CamelCase with no spaces.
	// For example, "ManagedProject".
	ReverseIdentifier string `json:"reverse_identifier,omitempty"`

	// The plural Go identifier that will be used for the reverse relationships.
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

package model

// Index will create accessor functions related to Columns.
type Index struct {
	// IsUnique indicates whether the index is unique
	IsUnique bool
	// Columns are the columns that are part of the index
	Columns    []*Column
	Name       string
	Identifier string
}

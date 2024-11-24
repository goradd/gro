package schema

// MultiColumnIndex declares that a LoadBy or QueryBy function should be created on the given columns in the table.
// If IsUnique is true, it will be a LoadBy function, otherwise a QueryBy. Databases that support multi-column indexes
// will have a matching multi-column index on the given columns.
type MultiColumnIndex struct {
	// Columns are the Column.Name values of the columns in the table that will be used to access the data.
	Columns []string `json:"columns"`
	// IsUnique will create a constraint in the database to make sure the combined columns are unique within
	// the table. Not all databases support uniqueness, but the generated ORM will attempt to check for existence
	// before adding the data to the database to minimize collisions in those databases that do not support uniqueness.
	IsUnique bool `json:"is_unique"`
}

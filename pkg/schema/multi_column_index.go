package schema

// MultiColumnIndex declares that a LoadBy or QueryBy function should be created on the given columns in the table.
//
// Databases that support multi-column indexes will have a matching multi-column index on the given columns.
//
// If IsUnique is true, a LoadBy function will be generated, otherwise a QueryBy function will be generated.
//
// Note On Uniqueness:
//
// Not all databases natively enforce uniqueness.
// The generated ORM will attempt to check for existence before adding the data to the database to minimize collisions
// but it is possible to still have a collision if two processes are trying to add the same value to two
// different records at the same time.
//
// The way to completely prevent it is to do one of the following:
//   - Use a database that enforces uniqueness, and then check for database errors specific to your
//     database that indicate a collision was detected during Save().
//   - Create a service that all instances of your application will use that implements a go channel
//     to lock values on columns during an insert or update, and that will prevent two clients from
//     obtaining a lock on the same value-column combinations. That service might need a timeout in case
//     the requester goes offline and never releases the lock, although leaving the lock just means the value
//     is permanently locked until the requester answers back, which might be fine.
//   - Whenever requesting a record by the unique value, you use a Query operation to detect if there are
//     duplicate records with the same value, and then respond accordingly if so.
type MultiColumnIndex struct {
	// Columns are the Column.Name values of the columns in the table that will be used to access the data.
	Columns []string `json:"columns"`
	// IsUnique will create a constraint in the database to make sure the combined columns are unique within
	// the table. Not all databases support uniqueness, but the generated ORM will attempt to check for existence
	// before adding the data to the database to minimize collisions in those databases that do not support uniqueness.
	IsUnique bool `json:"is_unique"`
}

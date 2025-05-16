package schema

// MultiColumnIndex declares that a Get or Load function should be created on the given columns in the table
// and gives direction to the database to create an index on the columns.
//
// Databases that support indexes will have a matching index on the given columns.
// See Column.IndexLevel to specify a single-column index.
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
	// IndexLevel will specify the type of index, and possibly a constraint to put on the columns.
	// If specifying a multi-column primary key, understand that the table cannot have any
	// columns marked as primary keys in the column definitions, that the table cannot have any foreign keys
	// pointing to it, and that some databases do not support multi-column primary keys in which case the driver
	// may set it up as a unique non-null index instead.
	IndexLevel IndexLevel `json:"index_level"`
}

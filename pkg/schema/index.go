package schema

import (
	"log/slog"
	"strings"
)

// Index declares that a Get or Load function should be created on the given columns in the table
// and gives direction to the database to create an index on the columns.
//
// Databases that support indexes will have a matching index on the given columns.
// See also Column.IndexLevel as another way to specify a single-column index.
//
// Note On Uniqueness:
//
// Not all databases natively enforce uniqueness.
// The generated ORM will attempt to check for existence before adding the data to the database to minimize collisions,
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
type Index struct {
	// Columns are the Column.Name values of the columns in the table that will be used to access the data.
	Columns []string `json:"columns"`
	// IndexLevel will specify the type of index, and possibly a constraint to put on the columns.
	IndexLevel IndexLevel `json:"index_level"`
	// Name is the name given to the index in the database, if the database requires a name.
	// The name will be concatenated with the name of the table and "idx".
	// Should be a snake_case word
	Name string `json:"name,omitempty"`
	// Identifier is the identifier used to create accessor functions.
	// Should be CamelCase.
	// A default will be generated from the column names if none is provided.
	Identifier string `json:"identifier,omitempty"`
}

func (i *Index) infer(t *Table) {
	if i.Name == "" {
		i.Name = t.Name + "_" + strings.Join(i.Columns, "_") + "_idx"
	}
}

func (i *Index) fillDefaults(t *Table) {
	if i.Identifier == "" {
		var colIDs []string
		for _, colName := range i.Columns {
			colID := t.columnIdentifier(colName)
			if colID == "" {
				slog.Error("Column not found",
					slog.String("column", colName),
					slog.String("table", t.Name))
			} else {
				colIDs = append(colIDs, colID)
			}
		}
		i.Identifier = strings.Join(colIDs, "")
	}
}

// For future expansion. Define a multi-column foreign key. The reference in those columns
// will need to point to a primary key column in the other table. Must use the Index structure
// to tell which columns make up a foreign key, since there is a possibility of two separate foreign keys
// pointing to the same table. (i.e. Mother and Father pointing to a person table).

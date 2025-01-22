package goradd_unit

// This is the implementation file for the DoubleIndex ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
)

// DoubleIndex represents an item in the double_index table in the database.
type DoubleIndex struct {
	doubleIndexBase
}

// NewDoubleIndex creates a new DoubleIndex object and initializes it to default values.
func NewDoubleIndex() *DoubleIndex {
	o := new(DoubleIndex)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a DoubleIndex database object to default values.
func (o *DoubleIndex) Initialize() {
	o.doubleIndexBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *DoubleIndex) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return fmt.Sprintf("DoubleIndex %v", o.PrimaryKey())
}

// QueryDoubleIndices returns a new query builder.
func QueryDoubleIndices(ctx context.Context) DoubleIndexBuilder {
	return queryDoubleIndices(ctx)
}

// queryDoubleIndices creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryDoubleIndices(ctx context.Context) DoubleIndexBuilder {
	return newDoubleIndexBuilder(ctx)
}

// DeleteDoubleIndex deletes a double_index record from the database given its primary key.
// Note that you can also delete loaded DoubleIndex objects by calling Delete on them.
// doc: type=DoubleIndex
func DeleteDoubleIndex(ctx context.Context, pk int) {
	deleteDoubleIndex(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitDoubleIndex", new(DoubleIndex))
}

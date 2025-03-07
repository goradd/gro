package goradd_unit

// This is the implementation file for the Root ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// Root represents an item in the root table in the database.
type Root struct {
	rootBase
}

// NewRoot creates a new Root object and initializes it to default values.
func NewRoot() *Root {
	o := new(Root)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a Root database object to default values.
func (o *Root) Initialize() {
	o.rootBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *Root) String() string {
	if o == nil {
		return ""
	}
	return o.name
}

// QueryRoots returns a new query builder.
func QueryRoots(ctx context.Context) RootBuilder {
	return queryRoots(ctx)
}

// queryRoots creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryRoots(ctx context.Context) RootBuilder {
	return newRootBuilder(ctx)
}

// DeleteRoot deletes the root record wtih primary key pk from the database.
// Note that you can also delete loaded Root objects by calling Delete on them.
// doc: type=Root
func DeleteRoot(ctx context.Context, pk string) {
	deleteRoot(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitRoot", new(Root))
}

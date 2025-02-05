package goradd_unit

// This is the implementation file for the ForwardNullUnique ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// ForwardNullUnique represents an item in the forward_null_unique table in the database.
type ForwardNullUnique struct {
	forwardNullUniqueBase
}

// NewForwardNullUnique creates a new ForwardNullUnique object and initializes it to default values.
func NewForwardNullUnique() *ForwardNullUnique {
	o := new(ForwardNullUnique)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a ForwardNullUnique database object to default values.
func (o *ForwardNullUnique) Initialize() {
	o.forwardNullUniqueBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *ForwardNullUnique) String() string {
	if o == nil {
		return ""
	}
	return o.name
}

// QueryForwardNullUniques returns a new query builder.
func QueryForwardNullUniques(ctx context.Context) ForwardNullUniqueBuilder {
	return queryForwardNullUniques(ctx)
}

// queryForwardNullUniques creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryForwardNullUniques(ctx context.Context) ForwardNullUniqueBuilder {
	return newForwardNullUniqueBuilder(ctx)
}

// DeleteForwardNullUnique deletes a forward_null_unique record from the database given its primary key.
// Note that you can also delete loaded ForwardNullUnique objects by calling Delete on them.
// doc: type=ForwardNullUnique
func DeleteForwardNullUnique(ctx context.Context, pk string) {
	deleteForwardNullUnique(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitForwardNullUnique", new(ForwardNullUnique))
}

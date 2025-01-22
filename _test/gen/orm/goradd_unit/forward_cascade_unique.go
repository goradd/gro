package goradd_unit

// This is the implementation file for the ForwardCascadeUnique ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// ForwardCascadeUnique represents an item in the forward_cascade_unique table in the database.
type ForwardCascadeUnique struct {
	forwardCascadeUniqueBase
}

// NewForwardCascadeUnique creates a new ForwardCascadeUnique object and initializes it to default values.
func NewForwardCascadeUnique() *ForwardCascadeUnique {
	o := new(ForwardCascadeUnique)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a ForwardCascadeUnique database object to default values.
func (o *ForwardCascadeUnique) Initialize() {
	o.forwardCascadeUniqueBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *ForwardCascadeUnique) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return o.name
}

// QueryForwardCascadeUniques returns a new query builder.
func QueryForwardCascadeUniques(ctx context.Context) ForwardCascadeUniqueBuilder {
	return queryForwardCascadeUniques(ctx)
}

// queryForwardCascadeUniques creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryForwardCascadeUniques(ctx context.Context) ForwardCascadeUniqueBuilder {
	return newForwardCascadeUniqueBuilder(ctx)
}

// DeleteForwardCascadeUnique deletes a forward_cascade_unique record from the database given its primary key.
// Note that you can also delete loaded ForwardCascadeUnique objects by calling Delete on them.
// doc: type=ForwardCascadeUnique
func DeleteForwardCascadeUnique(ctx context.Context, pk string) {
	deleteForwardCascadeUnique(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitForwardCascadeUnique", new(ForwardCascadeUnique))
}

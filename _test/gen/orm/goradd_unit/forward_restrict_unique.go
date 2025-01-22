package goradd_unit

// This is the implementation file for the ForwardRestrictUnique ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// ForwardRestrictUnique represents an item in the forward_restrict_unique table in the database.
type ForwardRestrictUnique struct {
	forwardRestrictUniqueBase
}

// NewForwardRestrictUnique creates a new ForwardRestrictUnique object and initializes it to default values.
func NewForwardRestrictUnique() *ForwardRestrictUnique {
	o := new(ForwardRestrictUnique)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a ForwardRestrictUnique database object to default values.
func (o *ForwardRestrictUnique) Initialize() {
	o.forwardRestrictUniqueBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *ForwardRestrictUnique) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return o.name
}

// QueryForwardRestrictUniques returns a new query builder.
func QueryForwardRestrictUniques(ctx context.Context) ForwardRestrictUniqueBuilder {
	return queryForwardRestrictUniques(ctx)
}

// queryForwardRestrictUniques creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryForwardRestrictUniques(ctx context.Context) ForwardRestrictUniqueBuilder {
	return newForwardRestrictUniqueBuilder(ctx)
}

// DeleteForwardRestrictUnique deletes a forward_restrict_unique record from the database given its primary key.
// Note that you can also delete loaded ForwardRestrictUnique objects by calling Delete on them.
// doc: type=ForwardRestrictUnique
func DeleteForwardRestrictUnique(ctx context.Context, pk string) {
	deleteForwardRestrictUnique(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitForwardRestrictUnique", new(ForwardRestrictUnique))
}

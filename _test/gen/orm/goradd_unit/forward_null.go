package goradd_unit

// This is the implementation file for the ForwardNull ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// ForwardNull represents an item in the forward_null table in the database.
type ForwardNull struct {
	forwardNullBase
}

// NewForwardNull creates a new ForwardNull object and initializes it to default values.
func NewForwardNull() *ForwardNull {
	o := new(ForwardNull)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a ForwardNull database object to default values.
func (o *ForwardNull) Initialize() {
	o.forwardNullBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *ForwardNull) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return o.name
}

// QueryForwardNulls returns a new query builder.
func QueryForwardNulls(ctx context.Context) ForwardNullBuilder {
	return queryForwardNulls(ctx)
}

// queryForwardNulls creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryForwardNulls(ctx context.Context) ForwardNullBuilder {
	return newForwardNullBuilder(ctx)
}

// DeleteForwardNull deletes a forward_null record from the database given its primary key.
// Note that you can also delete loaded ForwardNull objects by calling Delete on them.
// doc: type=ForwardNull
func DeleteForwardNull(ctx context.Context, pk string) {
	deleteForwardNull(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitForwardNull", new(ForwardNull))
}

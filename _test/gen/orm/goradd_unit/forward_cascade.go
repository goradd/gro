package goradd_unit

// This is the implementation file for the ForwardCascade ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// ForwardCascade represents an item in the forward_cascade table in the database.
type ForwardCascade struct {
	forwardCascadeBase
}

// NewForwardCascade creates a new ForwardCascade object and initializes it to default values.
func NewForwardCascade() *ForwardCascade {
	o := new(ForwardCascade)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a ForwardCascade database object to default values.
func (o *ForwardCascade) Initialize() {
	o.forwardCascadeBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *ForwardCascade) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return o.name
}

// QueryForwardCascades returns a new query builder.
func QueryForwardCascades(ctx context.Context) *ForwardCascadesBuilder {
	return queryForwardCascades(ctx)
}

// queryForwardCascades creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryForwardCascades(ctx context.Context) *ForwardCascadesBuilder {
	return newForwardCascadeBuilder(ctx)
}

// DeleteForwardCascade deletes a forward_cascade record from the database given its primary key.
// Note that you can also delete loaded ForwardCascade objects by calling Delete on them.
// doc: type=ForwardCascade
func DeleteForwardCascade(ctx context.Context, pk string) {
	deleteForwardCascade(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitForwardCascade", new(ForwardCascade))
}

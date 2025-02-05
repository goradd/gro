package goradd_unit

// This is the implementation file for the ForwardRestrict ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// ForwardRestrict represents an item in the forward_restrict table in the database.
type ForwardRestrict struct {
	forwardRestrictBase
}

// NewForwardRestrict creates a new ForwardRestrict object and initializes it to default values.
func NewForwardRestrict() *ForwardRestrict {
	o := new(ForwardRestrict)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a ForwardRestrict database object to default values.
func (o *ForwardRestrict) Initialize() {
	o.forwardRestrictBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *ForwardRestrict) String() string {
	if o == nil {
		return ""
	}
	return o.name
}

// QueryForwardRestricts returns a new query builder.
func QueryForwardRestricts(ctx context.Context) ForwardRestrictBuilder {
	return queryForwardRestricts(ctx)
}

// queryForwardRestricts creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryForwardRestricts(ctx context.Context) ForwardRestrictBuilder {
	return newForwardRestrictBuilder(ctx)
}

// DeleteForwardRestrict deletes a forward_restrict record from the database given its primary key.
// Note that you can also delete loaded ForwardRestrict objects by calling Delete on them.
// doc: type=ForwardRestrict
func DeleteForwardRestrict(ctx context.Context, pk string) {
	deleteForwardRestrict(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitForwardRestrict", new(ForwardRestrict))
}

package goradd_unit

// This is the implementation file for the Reverse ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// Reverse represents an item in the reverse table in the database.
type Reverse struct {
	reverseBase
}

// NewReverse creates a new Reverse object and initializes it to default values.
func NewReverse() *Reverse {
	o := new(Reverse)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a Reverse database object to default values.
func (o *Reverse) Initialize() {
	o.reverseBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *Reverse) String() string {
	if o == nil {
		return ""
	}
	return o.name
}

// QueryReverses returns a new query builder.
func QueryReverses(ctx context.Context) ReverseBuilder {
	return queryReverses(ctx)
}

// queryReverses creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryReverses(ctx context.Context) ReverseBuilder {
	return newReverseBuilder(ctx)
}

// DeleteReverse deletes a reverse record from the database given its primary key.
// Note that you can also delete loaded Reverse objects by calling Delete on them.
// doc: type=Reverse
func DeleteReverse(ctx context.Context, pk string) {
	deleteReverse(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitReverse", new(Reverse))
}

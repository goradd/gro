package goradd_unit

// This is the implementation file for the TwoKey ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
)

// TwoKey represents an item in the two_key table in the database.
type TwoKey struct {
	twoKeyBase
}

// NewTwoKey creates a new TwoKey object and initializes it to default values.
func NewTwoKey() *TwoKey {
	o := new(TwoKey)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a TwoKey database object to default values.
func (o *TwoKey) Initialize() {
	o.twoKeyBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *TwoKey) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return fmt.Sprintf("TwoKey %v", o.PrimaryKey())
}

// QueryTwoKeys returns a new query builder.
func QueryTwoKeys(ctx context.Context) *TwoKeysBuilder {
	return queryTwoKeys(ctx)
}

// queryTwoKeys creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryTwoKeys(ctx context.Context) *TwoKeysBuilder {
	return newTwoKeyBuilder(ctx)
}

func init() {
	gob.RegisterName("goradd_unitTwoKey", new(TwoKey))
}

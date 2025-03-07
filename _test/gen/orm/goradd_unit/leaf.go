package goradd_unit

// This is the implementation file for the Leaf ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// Leaf represents an item in the leaf table in the database.
type Leaf struct {
	leafBase
}

// NewLeaf creates a new Leaf object and initializes it to default values.
func NewLeaf() *Leaf {
	o := new(Leaf)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a Leaf database object to default values.
func (o *Leaf) Initialize() {
	o.leafBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *Leaf) String() string {
	if o == nil {
		return ""
	}
	return o.name
}

// QueryLeafs returns a new query builder.
func QueryLeafs(ctx context.Context) LeafBuilder {
	return queryLeafs(ctx)
}

// queryLeafs creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryLeafs(ctx context.Context) LeafBuilder {
	return newLeafBuilder(ctx)
}

// DeleteLeaf deletes the leaf record wtih primary key pk from the database.
// Note that you can also delete loaded Leaf objects by calling Delete on them.
// doc: type=Leaf
func DeleteLeaf(ctx context.Context, pk string) {
	deleteLeaf(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitLeaf", new(Leaf))
}

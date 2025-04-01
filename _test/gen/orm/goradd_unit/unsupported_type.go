package goradd_unit

// This is the implementation file for the UnsupportedType ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
)

// UnsupportedType represents an item in the unsupported_type table in the database.
type UnsupportedType struct {
	unsupportedTypeBase
}

// NewUnsupportedType creates a new UnsupportedType object and initializes it to default values.
func NewUnsupportedType() *UnsupportedType {
	o := new(UnsupportedType)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a UnsupportedType database object to default values.
func (o *UnsupportedType) Initialize() {
	o.unsupportedTypeBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns a description of the record, primarily for debugging.
func (o *UnsupportedType) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("UnsupportedType %v", o.PrimaryKey())
}

// Key returns a unique key for the object, among a list of similar objects.
func (o *UnsupportedType) Key() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("%v", o.PrimaryKey())
}

// Label returns a human-readable label of the object.
// This would be what a user would see as a description of the object if choosing from a list.
func (o *UnsupportedType) Label() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Unsupported Type %v", o.PrimaryKey())
}

// QueryUnsupportedTypes returns a new query builder.
func QueryUnsupportedTypes(ctx context.Context) UnsupportedTypeBuilder {
	return queryUnsupportedTypes(ctx)
}

// queryUnsupportedTypes creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryUnsupportedTypes(ctx context.Context) UnsupportedTypeBuilder {
	return newUnsupportedTypeBuilder(ctx)
}

// DeleteUnsupportedType deletes the unsupported_type record wtih primary key pk from the database.
// Note that you can also delete loaded UnsupportedType objects by calling Delete on them.
// doc: type=UnsupportedType
func DeleteUnsupportedType(ctx context.Context, pk string) {
	deleteUnsupportedType(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitUnsupportedType", new(UnsupportedType))
}

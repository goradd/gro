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

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *UnsupportedType) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("UnsupportedType %v", o.PrimaryKey())
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

// DeleteUnsupportedType deletes a unsupported_type record from the database given its primary key.
// Note that you can also delete loaded UnsupportedType objects by calling Delete on them.
// doc: type=UnsupportedType
func DeleteUnsupportedType(ctx context.Context, pk string) {
	deleteUnsupportedType(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitUnsupportedType", new(UnsupportedType))
}

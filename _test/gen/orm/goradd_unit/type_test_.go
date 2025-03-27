package goradd_unit

// This is the implementation file for the TypeTest ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
)

// TypeTest represents an item in the type_test table in the database.
type TypeTest struct {
	typeTestBase
}

// NewTypeTest creates a new TypeTest object and initializes it to default values.
func NewTypeTest() *TypeTest {
	o := new(TypeTest)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a TypeTest database object to default values.
func (o *TypeTest) Initialize() {
	o.typeTestBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns a description of the record, primarily for debugging.
func (o *TypeTest) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("TypeTest %v", o.PrimaryKey())
}

// Key returns a unique key for the object, among a list of similar objects.
func (o *TypeTest) Key() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("%v", o.PrimaryKey())
}

// Label returns a human readable label of the object.
// This would be what a user would see as a description of the object if choosing from a list.
func (o *TypeTest) Label() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Type Test %v", o.PrimaryKey())
}

// QueryTypeTests returns a new query builder.
func QueryTypeTests(ctx context.Context) TypeTestBuilder {
	return queryTypeTests(ctx)
}

// queryTypeTests creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryTypeTests(ctx context.Context) TypeTestBuilder {
	return newTypeTestBuilder(ctx)
}

// DeleteTypeTest deletes the type_test record wtih primary key pk from the database.
// Note that you can also delete loaded TypeTest objects by calling Delete on them.
// doc: type=TypeTest
func DeleteTypeTest(ctx context.Context, pk string) {
	deleteTypeTest(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitTypeTest", new(TypeTest))
}

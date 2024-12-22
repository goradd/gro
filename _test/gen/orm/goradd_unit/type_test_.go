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

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *TypeTest) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return fmt.Sprintf("TypeTest %v", o.PrimaryKey())
}

// QueryTypeTests returns a new query builder.
func QueryTypeTests(ctx context.Context) *TypeTestsBuilder {
	return queryTypeTests(ctx)
}

// queryTypeTests creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryTypeTests(ctx context.Context) *TypeTestsBuilder {
	return newTypeTestBuilder(ctx)
}

// DeleteTypeTest deletes a type_test record from the database given its primary key.
// Note that you can also delete loaded TypeTest objects by calling Delete on them.
// doc: type=TypeTest
func DeleteTypeTest(ctx context.Context, pk string) {
	deleteTypeTest(ctx, pk)
}

func init() {
	gob.RegisterName("goradd_unitTypeTest", new(TypeTest))
}

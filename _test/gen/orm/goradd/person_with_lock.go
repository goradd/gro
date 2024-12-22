package goradd

// This is the implementation file for the PersonWithLock ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
)

// PersonWithLock represents an item in the person_with_lock table in the database.
type PersonWithLock struct {
	personWithLockBase
}

// NewPersonWithLock creates a new PersonWithLock object and initializes it to default values.
func NewPersonWithLock() *PersonWithLock {
	o := new(PersonWithLock)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a PersonWithLock database object to default values.
func (o *PersonWithLock) Initialize() {
	o.personWithLockBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *PersonWithLock) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return fmt.Sprintf("PersonWithLock %v", o.PrimaryKey())
}

// QueryPersonWithLocks returns a new query builder.
func QueryPersonWithLocks(ctx context.Context) *PersonWithLocksBuilder {
	return queryPersonWithLocks(ctx)
}

// queryPersonWithLocks creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryPersonWithLocks(ctx context.Context) *PersonWithLocksBuilder {
	return newPersonWithLockBuilder(ctx)
}

// DeletePersonWithLock deletes a person_with_lock record from the database given its primary key.
// Note that you can also delete loaded PersonWithLock objects by calling Delete on them.
// doc: type=PersonWithLock
func DeletePersonWithLock(ctx context.Context, pk string) {
	deletePersonWithLock(ctx, pk)
}

func init() {
	gob.RegisterName("goraddPersonWithLock", new(PersonWithLock))
}

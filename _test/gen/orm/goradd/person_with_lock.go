package goradd

// This is the implementation file for the PersonWithLock ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"

	"github.com/goradd/gro/pkg/query"
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

// String implements the Stringer interface and returns a description of the record, primarily for debugging.
func (o *PersonWithLock) String() string {
	if o == nil {
		return ""
	}
	var pk string

	pk += fmt.Sprintf(" %v", o.ID())

	return "PersonWithLock" + pk
}

// Key returns a unique key for the object, among a list of similar objects.
func (o *PersonWithLock) Key() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("%v", o.PrimaryKey())
}

// Label returns a human-readable label of the object.
// This would be what a user would see as a description of the object if choosing from a list.
func (o *PersonWithLock) Label() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Person With Lock %v", o.PrimaryKey())
}

// Save will update or insert the object, depending on the state of the object.
//
// If it has an auto-generated primary key, it will be updated after an insert.
// Database errors generally will be handled by a panic and not returned here,
// since those indicate a problem with a database driver or configuration.
//
// Save will return a db.OptimisticLockError if it detects a collision when two users
// are attempting to change the same database record.
//
// It will return a db.UniqueValueError if it detects a collision when an attempt
// is made to add a record with a unique column that is given a value that is already in the database.
//
// Updating a record that has not changed will have no effect on the database.
// Updating a record that has linked records will also update any linked records that are MODIFIED,
// and if optimistic locking is in effect, will also check whether those records have been altered or deleted,
// returning an OptimisticLockError if so.
func (o *PersonWithLock) Save(ctx context.Context) error {
	return o.save(ctx)
}

// QueryPersonWithLocks returns a new query builder.
// See PersonWithLockBuilder for doc on how to use the builder.
// You should pass a context that has a timeout with it to protect against a long delay from
// the database possibly hanging your application. You can set a ReadTimeout value on the schema
// to do this by default during code generation.
func QueryPersonWithLocks(ctx context.Context) *PersonWithLockBuilder {
	return queryPersonWithLocks(ctx)
}

// queryPersonWithLocks creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryPersonWithLocks(ctx context.Context) *PersonWithLockBuilder {
	// Note: the context is provided here so that you can use it to enforce credentials if needed.
	// It is stored in the builder and later used in the terminating functions, like Load(), Get(), etc.
	// A QueryBuilder is meant to be a short-lived structure.
	return newPersonWithLockBuilder(ctx)
}

// getPersonWithLockInsertFields returns fields and values that will be used for a new record in the database.
// You can add or modify the fields here before they are sent to the database. If you set a primary key, it will be
// used instead of a generated primary key.
func getPersonWithLockInsertFields(o *personWithLockBase) (fields map[string]interface{}) {
	return o.getInsertFields()
}

// getPersonWithLockUpdateFields returns fields and values that will be used to update a current record in
// the database.
// You can add or modify the fields here before they are sent to the database.
func getPersonWithLockUpdateFields(o *personWithLockBase) (fields map[string]interface{}) {
	return o.getUpdateFields()
}

// DeletePersonWithLock deletes the person_with_lock record with primary key pk from the database.
// Note that you can also delete loaded PersonWithLock objects by calling Delete on them.
// Returns an error only if there was a problem with the database during the delete.
// If the record was not found, no error will be returned.
// doc: type=PersonWithLock
func DeletePersonWithLock(ctx context.Context, pk query.AutoPrimaryKey) error {
	return deletePersonWithLock(ctx, pk)
}

func init() {
	gob.RegisterName("goraddPersonWithLock", new(PersonWithLock))
}

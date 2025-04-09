package goradd

// This is the implementation file for the Person ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
)

// Person represents an item in the person table in the database.
type Person struct {
	personBase
}

// NewPerson creates a new Person object and initializes it to default values.
func NewPerson() *Person {
	o := new(Person)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a Person database object to default values.
func (o *Person) Initialize() {
	o.personBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns a description of the record, primarily for debugging.
func (o *Person) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Person %v", o.PrimaryKey())
}

// Key returns a unique key for the object, among a list of similar objects.
func (o *Person) Key() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("%v", o.PrimaryKey())
}

// Label returns a human-readable label of the object.
// This would be what a user would see as a description of the object if choosing from a list.
func (o *Person) Label() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Person %v", o.PrimaryKey())
}

// Save will update or insert the object, depending on the state of the object.
// If it has an auto-generated primary key, it will be changed after an insert.
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
func (o *Person) Save(ctx context.Context) error {
	return o.save(ctx)
}

// QueryPeople returns a new query builder.
func QueryPeople(ctx context.Context) PersonBuilder {
	return queryPeople(ctx)
}

// queryPeople creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryPeople(ctx context.Context) PersonBuilder {
	return newPersonBuilder(ctx)
}

// DeletePerson deletes the person record with primary key pk from the database.
// Note that you can also delete loaded Person objects by calling Delete on them.
// doc: type=Person
func DeletePerson(ctx context.Context, pk string) {
	deletePerson(ctx, pk)
}

func init() {
	gob.RegisterName("goraddPerson", new(Person))
}

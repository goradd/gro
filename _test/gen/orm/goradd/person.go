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

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *Person) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return fmt.Sprintf("Person %v", o.PrimaryKey())
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

// DeletePerson deletes a person record from the database given its primary key.
// Note that you can also delete loaded Person objects by calling Delete on them.
// doc: type=Person
func DeletePerson(ctx context.Context, pk string) {
	deletePerson(ctx, pk)
}

func init() {
	gob.RegisterName("goraddPerson", new(Person))
}

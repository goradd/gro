package goradd

// This is the implementation file for the Login ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
)

// Login represents an item in the login table in the database.
type Login struct {
	loginBase
}

// NewLogin creates a new Login object and initializes it to default values.
func NewLogin() *Login {
	o := new(Login)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a Login database object to default values.
func (o *Login) Initialize() {
	o.loginBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns a description of the record, primarily for debugging.
func (o *Login) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Login %v", o.PrimaryKey())
}

// Key returns a unique key for the object, among a list of similar objects.
func (o *Login) Key() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("%v", o.PrimaryKey())
}

// Label returns a human-readable label of the object.
// This would be what a user would see as a description of the object if choosing from a list.
func (o *Login) Label() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Login %v", o.PrimaryKey())
}

// QueryLogins returns a new query builder.
func QueryLogins(ctx context.Context) LoginBuilder {
	return queryLogins(ctx)
}

// queryLogins creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryLogins(ctx context.Context) LoginBuilder {
	return newLoginBuilder(ctx)
}

// DeleteLogin deletes the login record wtih primary key pk from the database.
// Note that you can also delete loaded Login objects by calling Delete on them.
// doc: type=Login
func DeleteLogin(ctx context.Context, pk string) {
	deleteLogin(ctx, pk)
}

func init() {
	gob.RegisterName("goraddLogin", new(Login))
}

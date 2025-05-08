package goradd

// This is the implementation file for the Address ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
	"fmt"
)

// Address represents an item in the address table in the database.
type Address struct {
	addressBase
}

// NewAddress creates a new Address object and initializes it to default values.
func NewAddress() *Address {
	o := new(Address)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a Address database object to default values.
func (o *Address) Initialize() {
	o.addressBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns a description of the record, primarily for debugging.
func (o *Address) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Address %v", o.PrimaryKey())
}

// Key returns a unique key for the object, among a list of similar objects.
func (o *Address) Key() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("%v", o.PrimaryKey())
}

// Label returns a human-readable label of the object.
// This would be what a user would see as a description of the object if choosing from a list.
func (o *Address) Label() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Address %v", o.PrimaryKey())
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
func (o *Address) Save(ctx context.Context) error {
	return o.save(ctx)
}

// QueryAddresses returns a new query builder.
func QueryAddresses(ctx context.Context) AddressBuilder {
	return queryAddresses(ctx)
}

// queryAddresses creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryAddresses(ctx context.Context) AddressBuilder {
	return newAddressBuilder(ctx)
}

// getAddressInsertFields returns fields and values that will be used for a new record in the database.
// You can add or modify the fields here before they are sent to the database. If you set a primary key, it will be
// used instead of a generated primary key.
func getAddressInsertFields(o *addressBase) (fields map[string]interface{}) {
	return o.getInsertFields()
}

// getAddressUpdateFields returns fields and values that will be used to update a current record in
// the database.
// You can add or modify the fields here before they are sent to the database.
func getAddressUpdateFields(o *addressBase) (fields map[string]interface{}) {
	return o.getUpdateFields()
}

// DeleteAddress deletes the address record with primary key pk from the database.
// Note that you can also delete loaded Address objects by calling Delete on them.
// Returns an error only if there was a problem with the database during the delete.
// If the record was not found, no error will be returned.
// doc: type=Address
func DeleteAddress(ctx context.Context, pk string) error {
	return deleteAddress(ctx, pk)
}

func init() {
	gob.RegisterName("goraddAddress", new(Address))
}

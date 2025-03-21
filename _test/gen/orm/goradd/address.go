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

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *Address) String() string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("Address %v", o.PrimaryKey())
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

// DeleteAddress deletes the address record wtih primary key pk from the database.
// Note that you can also delete loaded Address objects by calling Delete on them.
// doc: type=Address
func DeleteAddress(ctx context.Context, pk string) {
	deleteAddress(ctx, pk)
}

func init() {
	gob.RegisterName("goraddAddress", new(Address))
}

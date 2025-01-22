package goradd

// This is the implementation file for the Gift ORM object.
// This is where you build the api to your data model for your web application and potentially mobile apps.
// Your edits to this file will be preserved.

import (
	"context"
	"encoding/gob"
)

// Gift represents an item in the gift table in the database.
type Gift struct {
	giftBase
}

// NewGift creates a new Gift object and initializes it to default values.
func NewGift() *Gift {
	o := new(Gift)
	o.Initialize()
	return o
}

// Initialize will initialize or re-initialize a Gift database object to default values.
func (o *Gift) Initialize() {
	o.giftBase.Initialize()
	// Add your own initializations here
}

// String implements the Stringer interface and returns the default label for the object as it appears in html lists.
// Typically you would change this to whatever was pertinent to your application.
func (o *Gift) String() string {
	if o == nil {
		return "" // Possibly - Select One -?
	}
	return o.name
}

// QueryGifts returns a new query builder.
func QueryGifts(ctx context.Context) GiftBuilder {
	return queryGifts(ctx)
}

// queryGifts creates a new builder and is the central spot where all queries are directed.
// You can modify this function to enforce restrictions on queries, for example to make sure the user is authorized to
// access the data.
func queryGifts(ctx context.Context) GiftBuilder {
	return newGiftBuilder(ctx)
}

// DeleteGift deletes a gift record from the database given its primary key.
// Note that you can also delete loaded Gift objects by calling Delete on them.
// doc: type=Gift
func DeleteGift(ctx context.Context, pk int) {
	deleteGift(ctx, pk)
}

func init() {
	gob.RegisterName("goraddGift", new(Gift))
}

package goradd

// This is the test file for the Gift ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestGift_String(t *testing.T) {
	var obj *Gift

	assert.Equal(t, "", obj.String())

	obj = NewGift()
	s := obj.String()
	assert.Equal(t, "", s)
}

func TestGift_Delete(t *testing.T) {
	ctx := db.NewContext(nil)

	obj := createMinimalSampleGift()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteGift(ctx, obj.PrimaryKey())
	obj2 := LoadGift(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

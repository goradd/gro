package goradd

// This is the test file for the Gift ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGift_String(t *testing.T) {
	var obj *Gift

	assert.Equal(t, "", obj.String())

	obj = NewGift()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "Gift"))
}

func TestGift_Key(t *testing.T) {
	var obj *Gift
	assert.Equal(t, "", obj.Key())

	obj = NewGift()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestGift_Label(t *testing.T) {
	var obj *Gift
	assert.Equal(t, "", obj.Key())

	obj = NewGift()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestGift_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleGift()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteGift(ctx, obj.PrimaryKey()))
	obj2, err := LoadGift(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

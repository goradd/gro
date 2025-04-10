package goradd

// This is the test file for the Address ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestAddress_String(t *testing.T) {
	var obj *Address

	assert.Equal(t, "", obj.String())

	obj = NewAddress()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "Address"))
}

func TestAddress_Key(t *testing.T) {
	var obj *Address
	assert.Equal(t, "", obj.Key())

	obj = NewAddress()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestAddress_Label(t *testing.T) {
	var obj *Address
	assert.Equal(t, "", obj.Key())

	obj = NewAddress()
	s := obj.Label()
	assert.True(t, strings.HasPrefix(s, "Address"))
}

func TestAddress_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleAddress()
	assert.NoError(t, obj.Save(ctx))
	defer obj.Person().Delete(ctx)
	assert.NoError(t, DeleteAddress(ctx, obj.PrimaryKey()))
	obj2, err := LoadAddress(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

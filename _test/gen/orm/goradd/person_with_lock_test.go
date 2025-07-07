package goradd

// This is the test file for the PersonWithLock ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersonWithLock_String(t *testing.T) {
	var obj *PersonWithLock

	assert.Equal(t, "", obj.String())

	obj = NewPersonWithLock()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "PersonWithLock"))
}

func TestPersonWithLock_Key(t *testing.T) {
	var obj *PersonWithLock
	assert.Equal(t, "", obj.Key())

	obj = NewPersonWithLock()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestPersonWithLock_Label(t *testing.T) {
	var obj *PersonWithLock
	assert.Equal(t, "", obj.Key())

	obj = NewPersonWithLock()
	s := obj.Label()
	assert.True(t, strings.HasPrefix(s, "Person With Lock"))
}

func TestPersonWithLock_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSamplePersonWithLock()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeletePersonWithLock(ctx, obj.PrimaryKey()))
	obj2, err := LoadPersonWithLock(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

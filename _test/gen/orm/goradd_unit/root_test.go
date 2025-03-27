package goradd_unit

// This is the test file for the Root ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestRoot_String(t *testing.T) {
	var obj *Root

	assert.Equal(t, "", obj.String())

	obj = NewRoot()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "Root"))
}

func TestRoot_Key(t *testing.T) {
	var obj *Root
	assert.Equal(t, "", obj.Key())

	obj = NewRoot()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestRoot_Label(t *testing.T) {
	var obj *Root
	assert.Equal(t, "", obj.Key())

	obj = NewRoot()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestRoot_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleRoot()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteRoot(ctx, obj.PrimaryKey())
	obj2 := LoadRoot(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

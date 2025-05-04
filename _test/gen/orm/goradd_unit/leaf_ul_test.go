package goradd_unit

// This is the test file for the LeafUl ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestLeafUl_String(t *testing.T) {
	var obj *LeafUl

	assert.Equal(t, "", obj.String())

	obj = NewLeafUl()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "LeafUl"))
}

func TestLeafUl_Key(t *testing.T) {
	var obj *LeafUl
	assert.Equal(t, "", obj.Key())

	obj = NewLeafUl()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestLeafUl_Label(t *testing.T) {
	var obj *LeafUl
	assert.Equal(t, "", obj.Key())

	obj = NewLeafUl()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestLeafUl_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleLeafUl()
	assert.NoError(t, obj.Save(ctx))
	defer obj.RootUl().Delete(ctx)
	assert.NoError(t, DeleteLeafUl(ctx, obj.PrimaryKey()))
	obj2, err := LoadLeafUl(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

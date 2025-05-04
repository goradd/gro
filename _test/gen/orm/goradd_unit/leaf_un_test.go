package goradd_unit

// This is the test file for the LeafUn ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestLeafUn_String(t *testing.T) {
	var obj *LeafUn

	assert.Equal(t, "", obj.String())

	obj = NewLeafUn()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "LeafUn"))
}

func TestLeafUn_Key(t *testing.T) {
	var obj *LeafUn
	assert.Equal(t, "", obj.Key())

	obj = NewLeafUn()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestLeafUn_Label(t *testing.T) {
	var obj *LeafUn
	assert.Equal(t, "", obj.Key())

	obj = NewLeafUn()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestLeafUn_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleLeafUn()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteLeafUn(ctx, obj.PrimaryKey()))
	obj2, err := LoadLeafUn(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

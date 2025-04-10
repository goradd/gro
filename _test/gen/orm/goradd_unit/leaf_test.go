package goradd_unit

// This is the test file for the Leaf ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestLeaf_String(t *testing.T) {
	var obj *Leaf

	assert.Equal(t, "", obj.String())

	obj = NewLeaf()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "Leaf"))
}

func TestLeaf_Key(t *testing.T) {
	var obj *Leaf
	assert.Equal(t, "", obj.Key())

	obj = NewLeaf()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestLeaf_Label(t *testing.T) {
	var obj *Leaf
	assert.Equal(t, "", obj.Key())

	obj = NewLeaf()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestLeaf_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleLeaf()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteLeaf(ctx, obj.PrimaryKey()))
	obj2, err := LoadLeaf(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

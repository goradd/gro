package goradd_unit

// This is the test file for the LeafUnl ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestLeafUnl_String(t *testing.T) {
	var obj *LeafUnl

	assert.Equal(t, "", obj.String())

	obj = NewLeafUnl()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "LeafUnl"))
}

func TestLeafUnl_Key(t *testing.T) {
	var obj *LeafUnl
	assert.Equal(t, "", obj.Key())

	obj = NewLeafUnl()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestLeafUnl_Label(t *testing.T) {
	var obj *LeafUnl
	assert.Equal(t, "", obj.Key())

	obj = NewLeafUnl()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestLeafUnl_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleLeafUnl()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteLeafUnl(ctx, obj.PrimaryKey()))
	obj2, err := LoadLeafUnl(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

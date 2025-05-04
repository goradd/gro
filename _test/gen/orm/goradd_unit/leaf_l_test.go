package goradd_unit

// This is the test file for the LeafL ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestLeafL_String(t *testing.T) {
	var obj *LeafL

	assert.Equal(t, "", obj.String())

	obj = NewLeafL()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "LeafL"))
}

func TestLeafL_Key(t *testing.T) {
	var obj *LeafL
	assert.Equal(t, "", obj.Key())

	obj = NewLeafL()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestLeafL_Label(t *testing.T) {
	var obj *LeafL
	assert.Equal(t, "", obj.Key())

	obj = NewLeafL()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestLeafL_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleLeafL()
	assert.NoError(t, obj.Save(ctx))
	defer obj.RootL().Delete(ctx)
	assert.NoError(t, DeleteLeafL(ctx, obj.PrimaryKey()))
	obj2, err := LoadLeafL(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

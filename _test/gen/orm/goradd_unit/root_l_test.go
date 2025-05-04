package goradd_unit

// This is the test file for the RootL ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestRootL_String(t *testing.T) {
	var obj *RootL

	assert.Equal(t, "", obj.String())

	obj = NewRootL()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "RootL"))
}

func TestRootL_Key(t *testing.T) {
	var obj *RootL
	assert.Equal(t, "", obj.Key())

	obj = NewRootL()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestRootL_Label(t *testing.T) {
	var obj *RootL
	assert.Equal(t, "", obj.Key())

	obj = NewRootL()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestRootL_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleRootL()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteRootL(ctx, obj.PrimaryKey()))
	obj2, err := LoadRootL(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

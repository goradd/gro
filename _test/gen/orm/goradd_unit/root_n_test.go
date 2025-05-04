package goradd_unit

// This is the test file for the RootN ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestRootN_String(t *testing.T) {
	var obj *RootN

	assert.Equal(t, "", obj.String())

	obj = NewRootN()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "RootN"))
}

func TestRootN_Key(t *testing.T) {
	var obj *RootN
	assert.Equal(t, "", obj.Key())

	obj = NewRootN()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestRootN_Label(t *testing.T) {
	var obj *RootN
	assert.Equal(t, "", obj.Key())

	obj = NewRootN()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestRootN_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleRootN()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteRootN(ctx, obj.PrimaryKey()))
	obj2, err := LoadRootN(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

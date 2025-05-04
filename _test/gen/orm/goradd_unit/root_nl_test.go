package goradd_unit

// This is the test file for the RootNl ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestRootNl_String(t *testing.T) {
	var obj *RootNl

	assert.Equal(t, "", obj.String())

	obj = NewRootNl()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "RootNl"))
}

func TestRootNl_Key(t *testing.T) {
	var obj *RootNl
	assert.Equal(t, "", obj.Key())

	obj = NewRootNl()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestRootNl_Label(t *testing.T) {
	var obj *RootNl
	assert.Equal(t, "", obj.Key())

	obj = NewRootNl()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestRootNl_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleRootNl()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteRootNl(ctx, obj.PrimaryKey()))
	obj2, err := LoadRootNl(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

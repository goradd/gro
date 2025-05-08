package goradd_unit

// This is the test file for the AutoGen ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestAutoGen_String(t *testing.T) {
	var obj *AutoGen

	assert.Equal(t, "", obj.String())

	obj = NewAutoGen()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "AutoGen"))
}

func TestAutoGen_Key(t *testing.T) {
	var obj *AutoGen
	assert.Equal(t, "", obj.Key())

	obj = NewAutoGen()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestAutoGen_Label(t *testing.T) {
	var obj *AutoGen
	assert.Equal(t, "", obj.Key())

	obj = NewAutoGen()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestAutoGen_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleAutoGen()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteAutoGen(ctx, obj.PrimaryKey()))
	obj2, err := LoadAutoGen(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

package goradd_unit

// This is the test file for the RootU ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootU_String(t *testing.T) {
	var obj *RootU

	assert.Equal(t, "", obj.String())

	obj = NewRootU()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "RootU"))
}

func TestRootU_Key(t *testing.T) {
	var obj *RootU
	assert.Equal(t, "", obj.Key())

	obj = NewRootU()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestRootU_Label(t *testing.T) {
	var obj *RootU
	assert.Equal(t, "", obj.Key())

	obj = NewRootU()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestRootU_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleRootU()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteRootU(ctx, obj.PrimaryKey()))
	obj2, err := LoadRootU(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

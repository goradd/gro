package goradd_unit

// This is the test file for the RootUnl ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootUnl_String(t *testing.T) {
	var obj *RootUnl

	assert.Equal(t, "", obj.String())

	obj = NewRootUnl()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "RootUnl"))
}

func TestRootUnl_Key(t *testing.T) {
	var obj *RootUnl
	assert.Equal(t, "", obj.Key())

	obj = NewRootUnl()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestRootUnl_Label(t *testing.T) {
	var obj *RootUnl
	assert.Equal(t, "", obj.Key())

	obj = NewRootUnl()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestRootUnl_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleRootUnl()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteRootUnl(ctx, obj.PrimaryKey()))
	obj2, err := LoadRootUnl(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

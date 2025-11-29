package goradd_unit

// This is the test file for the RootUl ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootUl_String(t *testing.T) {
	var obj *RootUl

	assert.Equal(t, "", obj.String())

	obj = NewRootUl()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "RootUl"))
}

func TestRootUl_Key(t *testing.T) {
	var obj *RootUl
	assert.Equal(t, "", obj.Key())

	obj = NewRootUl()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestRootUl_Label(t *testing.T) {
	var obj *RootUl
	assert.Equal(t, "", obj.Key())

	obj = NewRootUl()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestRootUl_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleRootUl()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteRootUl(ctx, obj.PrimaryKey()))
	obj2, err := LoadRootUl(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

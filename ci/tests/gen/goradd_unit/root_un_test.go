package goradd_unit

// This is the test file for the RootUn ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootUn_String(t *testing.T) {
	var obj *RootUn

	assert.Equal(t, "", obj.String())

	obj = NewRootUn()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "RootUn"))
}

func TestRootUn_Key(t *testing.T) {
	var obj *RootUn
	assert.Equal(t, "", obj.Key())

	obj = NewRootUn()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestRootUn_Label(t *testing.T) {
	var obj *RootUn
	assert.Equal(t, "", obj.Key())

	obj = NewRootUn()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestRootUn_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleRootUn()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteRootUn(ctx, obj.PrimaryKey()))
	obj2, err := LoadRootUn(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

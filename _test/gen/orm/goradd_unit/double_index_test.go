package goradd_unit

// This is the test file for the DoubleIndex ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubleIndex_String(t *testing.T) {
	var obj *DoubleIndex

	assert.Equal(t, "", obj.String())

	obj = NewDoubleIndex()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "DoubleIndex"))
}

func TestDoubleIndex_Key(t *testing.T) {
	var obj *DoubleIndex
	assert.Equal(t, "", obj.Key())

	obj = NewDoubleIndex()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestDoubleIndex_Label(t *testing.T) {
	var obj *DoubleIndex
	assert.Equal(t, "", obj.Key())

	obj = NewDoubleIndex()
	s := obj.Label()
	assert.True(t, strings.HasPrefix(s, "Double Index"))
}

func TestDoubleIndex_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleDoubleIndex()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteDoubleIndex(ctx, obj.PrimaryKey()))
	obj2, err := LoadDoubleIndex(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

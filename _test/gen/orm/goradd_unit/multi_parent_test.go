package goradd_unit

// This is the test file for the MultiParent ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiParent_String(t *testing.T) {
	var obj *MultiParent

	assert.Equal(t, "", obj.String())

	obj = NewMultiParent()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "MultiParent"))
}

func TestMultiParent_Key(t *testing.T) {
	var obj *MultiParent
	assert.Equal(t, "", obj.Key())

	obj = NewMultiParent()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestMultiParent_Label(t *testing.T) {
	var obj *MultiParent
	assert.Equal(t, "", obj.Key())

	obj = NewMultiParent()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestMultiParent_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleMultiParent()
	assert.NoError(t, obj.Save(ctx))
	defer obj.Parent1().Delete(ctx)
	defer obj.Parent2().Delete(ctx)
	assert.NoError(t, DeleteMultiParent(ctx, obj.PrimaryKey()))
	obj2, err := LoadMultiParent(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

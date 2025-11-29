package goradd_unit

// This is the test file for the LeafN ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeafN_String(t *testing.T) {
	var obj *LeafN

	assert.Equal(t, "", obj.String())

	obj = NewLeafN()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "LeafN"))
}

func TestLeafN_Key(t *testing.T) {
	var obj *LeafN
	assert.Equal(t, "", obj.Key())

	obj = NewLeafN()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestLeafN_Label(t *testing.T) {
	var obj *LeafN
	assert.Equal(t, "", obj.Key())

	obj = NewLeafN()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestLeafN_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleLeafN()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteLeafN(ctx, obj.PrimaryKey()))
	obj2, err := LoadLeafN(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

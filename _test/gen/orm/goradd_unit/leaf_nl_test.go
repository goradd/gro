package goradd_unit

// This is the test file for the LeafNl ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeafNl_String(t *testing.T) {
	var obj *LeafNl

	assert.Equal(t, "", obj.String())

	obj = NewLeafNl()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "LeafNl"))
}

func TestLeafNl_Key(t *testing.T) {
	var obj *LeafNl
	assert.Equal(t, "", obj.Key())

	obj = NewLeafNl()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestLeafNl_Label(t *testing.T) {
	var obj *LeafNl
	assert.Equal(t, "", obj.Key())

	obj = NewLeafNl()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestLeafNl_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleLeafNl()
	assert.NoError(t, obj.Save(ctx))
	defer obj.RootNl().Delete(ctx)
	assert.NoError(t, DeleteLeafNl(ctx, obj.PrimaryKey()))
	obj2, err := LoadLeafNl(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

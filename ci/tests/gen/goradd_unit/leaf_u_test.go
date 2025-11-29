package goradd_unit

// This is the test file for the LeafU ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeafU_String(t *testing.T) {
	var obj *LeafU

	assert.Equal(t, "", obj.String())

	obj = NewLeafU()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "LeafU"))
}

func TestLeafU_Key(t *testing.T) {
	var obj *LeafU
	assert.Equal(t, "", obj.Key())

	obj = NewLeafU()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestLeafU_Label(t *testing.T) {
	var obj *LeafU
	assert.Equal(t, "", obj.Key())

	obj = NewLeafU()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestLeafU_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleLeafU()
	assert.NoError(t, obj.Save(ctx))
	defer obj.RootU().Delete(ctx)
	assert.NoError(t, DeleteLeafU(ctx, obj.PrimaryKey()))
	obj2, err := LoadLeafU(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

package goradd_unit

// This is the test file for the TwoKey ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTwoKey_String(t *testing.T) {
	var obj *TwoKey

	assert.Equal(t, "", obj.String())

	obj = NewTwoKey()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "TwoKey"))
}

func TestTwoKey_Key(t *testing.T) {
	var obj *TwoKey
	assert.Equal(t, "", obj.Key())

	obj = NewTwoKey()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestTwoKey_Label(t *testing.T) {
	var obj *TwoKey
	assert.Equal(t, "", obj.Key())

	obj = NewTwoKey()
	s := obj.Label()
	assert.True(t, strings.HasPrefix(s, "Two Key"))
}

func TestTwoKey_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleTwoKey()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteTwoKey(ctx, obj.PrimaryKey()))
	obj2, err := LoadTwoKey(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

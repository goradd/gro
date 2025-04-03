package goradd_unit

// This is the test file for the MultiParent ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
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
	ctx := db.NewContext(nil)
	obj := createMinimalSampleMultiParent()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteMultiParent(ctx, obj.PrimaryKey())
	obj2 := LoadMultiParent(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

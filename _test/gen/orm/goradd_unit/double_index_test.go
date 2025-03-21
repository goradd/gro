package goradd_unit

// This is the test file for the DoubleIndex ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestDoubleIndex_String(t *testing.T) {
	var obj *DoubleIndex

	assert.Equal(t, "", obj.String())

	obj = NewDoubleIndex()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "DoubleIndex"))
}

func TestDoubleIndex_Delete(t *testing.T) {
	ctx := db.NewContext(nil)

	obj := createMinimalSampleDoubleIndex()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteDoubleIndex(ctx, obj.PrimaryKey())
	obj2 := LoadDoubleIndex(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

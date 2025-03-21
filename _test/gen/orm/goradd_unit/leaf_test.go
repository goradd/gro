package goradd_unit

// This is the test file for the Leaf ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestLeaf_String(t *testing.T) {
	var obj *Leaf

	assert.Equal(t, "", obj.String())

	obj = NewLeaf()
	s := obj.String()
	assert.Equal(t, "", s)
}

func TestLeaf_Delete(t *testing.T) {
	ctx := db.NewContext(nil)

	obj := createMinimalSampleLeaf()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteLeaf(ctx, obj.PrimaryKey())
	obj2 := LoadLeaf(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

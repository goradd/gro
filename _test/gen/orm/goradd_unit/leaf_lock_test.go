package goradd_unit

// This is the test file for the LeafLock ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestLeafLock_String(t *testing.T) {
	var obj *LeafLock

	assert.Equal(t, "", obj.String())

	obj = NewLeafLock()
	s := obj.String()
	assert.Equal(t, "", s)
}

func TestLeafLock_Delete(t *testing.T) {
	ctx := db.NewContext(nil)

	obj := createMinimalSampleLeafLock()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteLeafLock(ctx, obj.PrimaryKey())
	obj2 := LoadLeafLock(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

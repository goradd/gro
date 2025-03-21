package goradd

// This is the test file for the Address ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestAddress_String(t *testing.T) {
	var obj *Address

	assert.Equal(t, "", obj.String())

	obj = NewAddress()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "Address"))
}

func TestAddress_Delete(t *testing.T) {
	ctx := db.NewContext(nil)

	obj := createMinimalSampleAddress()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteAddress(ctx, obj.PrimaryKey())
	obj2 := LoadAddress(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

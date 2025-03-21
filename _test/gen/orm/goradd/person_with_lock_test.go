package goradd

// This is the test file for the PersonWithLock ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestPersonWithLock_String(t *testing.T) {
	var obj *PersonWithLock

	assert.Equal(t, "", obj.String())

	obj = NewPersonWithLock()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "PersonWithLock"))
}

func TestPersonWithLock_Delete(t *testing.T) {
	ctx := db.NewContext(nil)

	obj := createMinimalSamplePersonWithLock()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeletePersonWithLock(ctx, obj.PrimaryKey())
	obj2 := LoadPersonWithLock(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

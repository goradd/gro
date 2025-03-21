package goradd_unit

// This is the test file for the TypeTest ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestTypeTest_String(t *testing.T) {
	var obj *TypeTest

	assert.Equal(t, "", obj.String())

	obj = NewTypeTest()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "TypeTest"))
}

func TestTypeTest_Delete(t *testing.T) {
	ctx := db.NewContext(nil)

	obj := createMinimalSampleTypeTest()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteTypeTest(ctx, obj.PrimaryKey())
	obj2 := LoadTypeTest(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

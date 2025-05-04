package goradd_unit

// This is the test file for the TypeTest ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
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

func TestTypeTest_Key(t *testing.T) {
	var obj *TypeTest
	assert.Equal(t, "", obj.Key())

	obj = NewTypeTest()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestTypeTest_Label(t *testing.T) {
	var obj *TypeTest
	assert.Equal(t, "", obj.Key())

	obj = NewTypeTest()
	s := obj.Label()
	assert.True(t, strings.HasPrefix(s, "Type Test"))
}

func TestTypeTest_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleTypeTest()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteTypeTest(ctx, obj.PrimaryKey()))
	obj2, err := LoadTypeTest(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

package goradd

// This is the test file for the Person ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestPerson_String(t *testing.T) {
	var obj *Person

	assert.Equal(t, "", obj.String())

	obj = NewPerson()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "Person"))
}

func TestPerson_Key(t *testing.T) {
	var obj *Person
	assert.Equal(t, "", obj.Key())

	obj = NewPerson()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestPerson_Label(t *testing.T) {
	var obj *Person
	assert.Equal(t, "", obj.Key())

	obj = NewPerson()
	s := obj.Label()
	assert.True(t, strings.HasPrefix(s, "Person"))
}

func TestPerson_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSamplePerson()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeletePerson(ctx, obj.PrimaryKey())
	obj2 := LoadPerson(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

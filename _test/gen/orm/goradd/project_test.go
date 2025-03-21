package goradd

// This is the test file for the Project ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestProject_String(t *testing.T) {
	var obj *Project

	assert.Equal(t, "", obj.String())

	obj = NewProject()
	s := obj.String()
	assert.Equal(t, "", s)
}

func TestProject_Delete(t *testing.T) {
	ctx := db.NewContext(nil)

	obj := createMinimalSampleProject()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteProject(ctx, obj.PrimaryKey())
	obj2 := LoadProject(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

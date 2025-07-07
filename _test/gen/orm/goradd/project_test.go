package goradd

// This is the test file for the Project ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject_String(t *testing.T) {
	var obj *Project

	assert.Equal(t, "", obj.String())

	obj = NewProject()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "Project"))
}

func TestProject_Key(t *testing.T) {
	var obj *Project
	assert.Equal(t, "", obj.Key())

	obj = NewProject()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestProject_Label(t *testing.T) {
	var obj *Project
	assert.Equal(t, "", obj.Key())

	obj = NewProject()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestProject_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleProject()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteProject(ctx, obj.PrimaryKey()))
	obj2, err := LoadProject(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

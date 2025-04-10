package goradd

// This is the test file for the EmployeeInfo ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestEmployeeInfo_String(t *testing.T) {
	var obj *EmployeeInfo

	assert.Equal(t, "", obj.String())

	obj = NewEmployeeInfo()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "EmployeeInfo"))
}

func TestEmployeeInfo_Key(t *testing.T) {
	var obj *EmployeeInfo
	assert.Equal(t, "", obj.Key())

	obj = NewEmployeeInfo()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestEmployeeInfo_Label(t *testing.T) {
	var obj *EmployeeInfo
	assert.Equal(t, "", obj.Key())

	obj = NewEmployeeInfo()
	s := obj.Label()
	assert.True(t, strings.HasPrefix(s, "Employee Info"))
}

func TestEmployeeInfo_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleEmployeeInfo()
	assert.NoError(t, obj.Save(ctx))
	defer obj.Person().Delete(ctx)
	assert.NoError(t, DeleteEmployeeInfo(ctx, obj.PrimaryKey()))
	obj2, err := LoadEmployeeInfo(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

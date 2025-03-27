package goradd

// This is the test file for the Milestone ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestMilestone_String(t *testing.T) {
	var obj *Milestone

	assert.Equal(t, "", obj.String())

	obj = NewMilestone()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "Milestone"))
}

func TestMilestone_Key(t *testing.T) {
	var obj *Milestone
	assert.Equal(t, "", obj.Key())

	obj = NewMilestone()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestMilestone_Label(t *testing.T) {
	var obj *Milestone
	assert.Equal(t, "", obj.Key())

	obj = NewMilestone()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestMilestone_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleMilestone()
	err := obj.Save(ctx)
	defer obj.Project().Delete(ctx)
	assert.NoError(t, err)
	DeleteMilestone(ctx, obj.PrimaryKey())
	obj2 := LoadMilestone(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

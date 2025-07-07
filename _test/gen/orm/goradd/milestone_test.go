package goradd

// This is the test file for the Milestone ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

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
}

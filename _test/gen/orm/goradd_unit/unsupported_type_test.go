package goradd_unit

// This is the test file for the UnsupportedType ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsupportedType_String(t *testing.T) {
	var obj *UnsupportedType

	assert.Equal(t, "", obj.String())

	obj = NewUnsupportedType()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "UnsupportedType"))
}

func TestUnsupportedType_Delete(t *testing.T) {
}

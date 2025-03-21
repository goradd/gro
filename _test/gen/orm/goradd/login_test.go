package goradd

// This is the test file for the Login ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestLogin_String(t *testing.T) {
	var obj *Login

	assert.Equal(t, "", obj.String())

	obj = NewLogin()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "Login"))
}

func TestLogin_Delete(t *testing.T) {
	ctx := db.NewContext(nil)

	obj := createMinimalSampleLogin()
	err := obj.Save(ctx)
	assert.NoError(t, err)
	DeleteLogin(ctx, obj.PrimaryKey())
	obj2 := LoadLogin(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
}

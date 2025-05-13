package goradd_unit

// This is the test file for the AltRootUn ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goradd/orm/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestAltRootUn_String(t *testing.T) {
	var obj *AltRootUn

	assert.Equal(t, "", obj.String())

	obj = NewAltRootUn()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "AltRootUn"))
}

func TestAltRootUn_Key(t *testing.T) {
	var obj *AltRootUn
	assert.Equal(t, "", obj.Key())

	obj = NewAltRootUn()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestAltRootUn_Label(t *testing.T) {
	var obj *AltRootUn
	assert.Equal(t, "", obj.Key())

	obj = NewAltRootUn()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestAltRootUn_Delete(t *testing.T) {
	ctx := db.NewContext(nil)
	obj := createMinimalSampleAltRootUn()
	assert.NoError(t, obj.Save(ctx))
	assert.NoError(t, DeleteAltRootUn(ctx, obj.PrimaryKey()))
	obj2, err := LoadAltRootUn(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

package goradd_unit

// This is the test file for the AltLeafUn ORM object.
// Add your tests to this file or modify the one provided.
// Your edits to this file will be preserved.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAltLeafUn_String(t *testing.T) {
	var obj *AltLeafUn

	assert.Equal(t, "", obj.String())

	obj = NewAltLeafUn()
	s := obj.String()
	assert.True(t, strings.HasPrefix(s, "AltLeafUn"))
}

func TestAltLeafUn_Key(t *testing.T) {
	var obj *AltLeafUn
	assert.Equal(t, "", obj.Key())

	obj = NewAltLeafUn()
	assert.Equal(t, fmt.Sprintf("%v", obj.PrimaryKey()), obj.Key())
}

func TestAltLeafUn_Label(t *testing.T) {
	var obj *AltLeafUn
	assert.Equal(t, "", obj.Key())

	obj = NewAltLeafUn()
	s := obj.Label()
	assert.Equal(t, "", s)
}

func TestAltLeafUn_Delete(t *testing.T) {
	ctx := context.Background()
	obj := createMinimalSampleAltLeafUn()
	assert.NoError(t, obj.Save(ctx))
	defer obj.AltRootUn().Delete(ctx)
	assert.NoError(t, DeleteAltLeafUn(ctx, obj.PrimaryKey()))
	obj2, err := LoadAltLeafUn(ctx, obj.PrimaryKey())
	assert.Nil(t, obj2)
	assert.NoError(t, err)
}

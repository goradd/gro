package goradd

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestAddress_String(t *testing.T) {
	a := NewAddress()
	s := a.String()
	assert.True(t, strings.HasPrefix(s, "Address"))
}

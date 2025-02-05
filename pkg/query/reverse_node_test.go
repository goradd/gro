package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReverseReferenceNodeInterfaces(t *testing.T) {
	n := &ReverseNode{
		ColumnQueryName: "col",
		Identifier:      "Obj",
		ReceiverType:    ColTypeString,
	}
	
	assert.Implements(t, (*linker)(nil), n)
}
